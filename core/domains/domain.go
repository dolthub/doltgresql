// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package domains

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// Domain represents a single domain.
type Domain struct {
	Name        string
	DataType    types.DoltgresType
	DefaultExpr string
	NotNull     bool
	Checks      []*sql.CheckDefinition
}

// NewDomain creates new instance of Domain.
func NewDomain(ctx *sql.Context, name string, typ types.DoltgresType, defVal sql.Expression, notNull bool, checks sql.CheckConstraints) (*Domain, error) {
	checkDefs := make([]*sql.CheckDefinition, len(checks))
	var defExpr string
	if defVal != nil {
		defExpr = defVal.String()
	}
	var err error
	for i, check := range checks {
		checkDefs[i], err = plan.NewCheckDefinition(ctx, check)
		if err != nil {
			return nil, err
		}
	}
	return &Domain{
		Name:        name,
		DataType:    typ,
		DefaultExpr: defExpr,
		NotNull:     notNull,
		Checks:      checkDefs,
	}, nil
}

// DomainCollection contains a collection of domains.
type DomainCollection struct {
	schemaMap map[string]map[string]*Domain
	mutex     *sync.Mutex
}

// GetDomain returns the domain with the given schema and name.
// Returns nil if the domain cannot be found.
func (pgs *DomainCollection) GetDomain(schName, domainName string) (*Domain, bool) {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if nameMap, ok := pgs.schemaMap[schName]; ok {
		if domain, ok := nameMap[domainName]; ok {
			return domain, true
		}
	}
	return nil, false
}

// GetAllDomains returns a map containing all domains in the collection, grouped by the schema they're contained in.
// Each domain array is also sorted by the domain name.
func (pgs *DomainCollection) GetAllDomains() (domainsMap map[string][]*Domain, schemaNames []string, totalCount int) {
	domainsMap = make(map[string][]*Domain)
	for schemaName, nameMap := range pgs.schemaMap {
		schemaNames = append(schemaNames, schemaName)
		domains := make([]*Domain, 0, len(nameMap))
		for _, domain := range nameMap {
			domains = append(domains, domain)
		}
		totalCount += len(domains)
		sort.Slice(domains, func(i, j int) bool {
			return domains[i].Name < domains[j].Name
		})
		domainsMap[schemaName] = domains
	}
	sort.Slice(schemaNames, func(i, j int) bool {
		return schemaNames[i] < schemaNames[j]
	})
	return
}

// CreateDomain creates a new domain.
func (pgs *DomainCollection) CreateDomain(schema string, domain *Domain) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	nameMap, ok := pgs.schemaMap[schema]
	if !ok {
		nameMap = make(map[string]*Domain)
		pgs.schemaMap[schema] = nameMap
	}
	if _, ok = nameMap[domain.Name]; ok {
		return types.ErrTypeAlreadyExists.New(domain.Name)
	}
	nameMap[domain.Name] = domain
	return nil
}

// DropDomain drops an existing domain.
func (pgs *DomainCollection) DropDomain(schName, domainName string) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if nameMap, ok := pgs.schemaMap[schName]; ok {
		if _, ok = nameMap[domainName]; ok {
			delete(nameMap, domainName)
			return nil
		}
	}
	return types.ErrTypeDoesNotExist.New(domainName)
}

// IterateDomains iterates over all domains in the collection.
func (pgs *DomainCollection) IterateDomains(f func(schema string, domain *Domain) error) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	for schema, nameMap := range pgs.schemaMap {
		for _, domain := range nameMap {
			if err := f(schema, domain); err != nil {
				return err
			}
		}
	}
	return nil
}

// Clone returns a new *DomainCollection with the same contents as the original.
func (pgs *DomainCollection) Clone() *DomainCollection {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	newCollection := &DomainCollection{
		schemaMap: make(map[string]map[string]*Domain),
		mutex:     &sync.Mutex{},
	}
	for schema, nameMap := range pgs.schemaMap {
		if len(nameMap) == 0 {
			continue
		}
		clonedNameMap := make(map[string]*Domain)
		for key, domain := range nameMap {
			newDomain := *domain
			clonedNameMap[key] = &newDomain
		}
		newCollection.schemaMap[schema] = clonedNameMap
	}
	return newCollection
}

// Serialize returns the DomainCollection as a byte slice.
// If the DomainCollection is nil, then this returns a nil slice.
func (pgs *DomainCollection) Serialize(ctx context.Context) ([]byte, error) {
	if pgs == nil {
		return nil, nil
	}
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// Write all the domains to the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	schemaMapKeys := utils.GetMapKeysSorted(pgs.schemaMap)
	writer.VariableUint(uint64(len(schemaMapKeys)))
	for _, schemaMapKey := range schemaMapKeys {
		nameMap := pgs.schemaMap[schemaMapKey]
		writer.String(schemaMapKey)
		nameMapKeys := utils.GetMapKeysSorted(nameMap)
		writer.VariableUint(uint64(len(nameMapKeys)))
		for _, nameMapKey := range nameMapKeys {
			domain := nameMap[nameMapKey]
			writer.String(domain.Name)
			writer.Uint32(domain.DataType.OID())
			writer.String(domain.DefaultExpr)
			writer.Bool(domain.NotNull)
			writer.VariableUint(uint64(len(domain.Checks)))
			for _, check := range domain.Checks {
				writer.String(check.Name)
				writer.String(check.CheckExpression)
			}
		}
	}

	return writer.Data(), nil
}

// Deserialize returns the Collection that was serialized in the byte slice.
// Returns an empty Collection if data is nil or empty.
func Deserialize(ctx context.Context, data []byte) (*DomainCollection, error) {
	if len(data) == 0 {
		return &DomainCollection{
			schemaMap: make(map[string]map[string]*Domain),
			mutex:     &sync.Mutex{},
		}, nil
	}
	schemaMap := make(map[string]map[string]*Domain)
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return nil, fmt.Errorf("version %d of domains is not supported, please upgrade the server", version)
	}

	// Read from the reader
	numOfSchemas := reader.VariableUint()
	for i := uint64(0); i < numOfSchemas; i++ {
		schemaName := reader.String()
		numOfDomains := reader.VariableUint()
		nameMap := make(map[string]*Domain)
		for j := uint64(0); j < numOfDomains; j++ {
			domain := &Domain{}
			domain.Name = reader.String()
			domain.DataType = types.OidToBuildInDoltgresType[reader.Uint32()]
			domain.DefaultExpr = reader.String()
			domain.NotNull = reader.Bool()
			numOfChecks := reader.VariableUint()
			for k := uint64(0); k < numOfChecks; k++ {
				checkName := reader.String()
				checkExpr := reader.String()
				domain.Checks = append(domain.Checks, &sql.CheckDefinition{
					Name:            checkName,
					CheckExpression: checkExpr,
					Enforced:        true,
				})
			}
			nameMap[domain.Name] = domain
		}
		schemaMap[schemaName] = nameMap
	}
	if !reader.IsEmpty() {
		return nil, fmt.Errorf("extra data found while deserializing domains")
	}

	// Return the deserialized object
	return &DomainCollection{
		schemaMap: schemaMap,
		mutex:     &sync.Mutex{},
	}, nil
}
