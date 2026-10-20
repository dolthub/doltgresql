// Copyright 2026 Dolthub, Inc.
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

package v3_6_4

import (
	"fmt"
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// partitionGeometries holds the geometries of a window partition for clustering, where `present` marks the rows that
// hold a geometry that is not empty.
type partitionGeometries struct {
	geos    *spatial.GEOS
	group   *spatial.Group
	boxes   []bbox
	present []bool
	indexes []int
}

// newPartitionGeometries converts the geometries in the first argument of each row for clustering. The caller must
// close the result.
func newPartitionGeometries(rows [][]any) (*partitionGeometries, error) {
	p := &partitionGeometries{geos: spatial.NewGEOS(), boxes: make([]bbox, len(rows)), present: make([]bool, len(rows)), indexes: make([]int, len(rows))}
	var converted []spatial.Geometry
	for i, row := range rows {
		g, ok := row[0].(Geometry)
		if !ok || g.Shape.isEmpty() {
			continue
		}
		geosG, err := g.Shape.toGEOS()
		if err != nil {
			p.geos.Close()
			return nil, err
		}
		p.boxes[i], _ = g.Shape.bounds()
		p.present[i], p.indexes[i] = true, len(converted)
		converted = append(converted, geosG)
	}
	group, err := p.geos.NewGroup(converted)
	if err != nil {
		p.geos.Close()
		return nil, geosError("LWGEOM2GEOS", err)
	}
	p.group = group
	return p, nil
}

// close releases the partition's geometries.
func (p *partitionGeometries) close() {
	p.group.Close()
	p.geos.Close()
}

// withinDistance returns whether the geometries of two rows lie within the distance of each other.
func (p *partitionGeometries) withinDistance(i int, j int, distance float64) (bool, error) {
	a, b := p.boxes[i], p.boxes[j]
	if math.Max(a.minX-b.maxX, b.minX-a.maxX) > distance || math.Max(a.minY-b.maxY, b.minY-a.maxY) > distance {
		return false, nil
	}
	d, err := p.group.Distance(p.indexes[i], p.indexes[j])
	if err != nil {
		return false, geosError("GEOSDistance", err)
	}
	return d <= distance, nil
}

// intersects returns whether the geometries of two rows share a point.
func (p *partitionGeometries) intersects(i int, j int) (bool, error) {
	a, b := p.boxes[i], p.boxes[j]
	if a.minX > b.maxX || b.minX > a.maxX || a.minY > b.maxY || b.minY > a.maxY {
		return false, nil
	}
	intersects, err := p.group.Intersects(p.indexes[i], p.indexes[j])
	if err != nil {
		return false, geosError("GEOSIntersects", err)
	}
	return intersects, nil
}

// unionFind joins rows into clusters.
type unionFind []int

// newUnionFind returns a union-find where every one of the given number of rows is its own cluster.
func newUnionFind(count int) unionFind {
	u := make(unionFind, count)
	for i := range u {
		u[i] = i
	}
	return u
}

// find returns the row that represents the cluster of a row.
func (u unionFind) find(i int) int {
	for u[i] != i {
		u[i] = u[u[i]]
		i = u[i]
	}
	return i
}

// union joins the clusters of two rows.
func (u unionFind) union(i int, j int) {
	if a, b := u.find(i), u.find(j); a != b {
		u[b] = a
	}
}

// clusterNumbers returns each row's cluster number, numbering the clusters in the order of their first rows. Rows
// that `skip` reports take no number and get null, and the rows that `hide` reports use up a number but get null.
func clusterNumbers(u unionFind, skip func(i int) bool, hide func(i int) bool) []any {
	numbers := make(map[int]int32)
	results := make([]any, len(u))
	for i := range u {
		if skip(i) {
			continue
		}
		root := u.find(i)
		number, ok := numbers[root]
		if !ok {
			number = int32(len(numbers))
			numbers[root] = number
		}
		if !hide(i) {
			results[i] = number
		}
	}
	return results
}

// clusterWindow returns the implementation of ST_ClusterWithinWin or ST_ClusterIntersectingWin, which join rows whose
// geometries are near each other directly or through other rows. A row without a geometry is its own cluster with a
// null number.
func clusterWindow(near func(p *partitionGeometries, i int, j int, row []any) (bool, error)) extdef.WindowFunction {
	return func(ctx *sql.Context, rows [][]any) ([]any, error) {
		p, err := newPartitionGeometries(rows)
		if err != nil {
			return nil, err
		}
		defer p.close()
		u := newUnionFind(len(rows))
		for i := range rows {
			for j := i + 1; j < len(rows) && p.present[i]; j++ {
				if !p.present[j] || u.find(i) == u.find(j) {
					continue
				}
				isNear, err := near(p, i, j, rows[i])
				if err != nil {
					return nil, err
				}
				if isNear {
					u.union(i, j)
				}
			}
		}
		return clusterNumbers(u, func(int) bool { return false }, func(i int) bool { return rows[i][0] == nil }), nil
	}
}

// clusterDBSCAN implements ST_ClusterDBSCAN, which clusters the rows whose geometries have at least `minpoints`
// geometries within `eps` of them, counting themselves, together with the geometries within `eps` of those. Every
// other row is noise, with a null number.
func clusterDBSCAN(ctx *sql.Context, rows [][]any) ([]any, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	eps, ok := rows[0][1].(float64)
	minPoints, ok2 := rows[0][2].(int32)
	if !ok || !ok2 {
		return make([]any, len(rows)), nil
	}
	p, err := newPartitionGeometries(rows)
	if err != nil {
		return nil, err
	}
	defer p.close()
	neighbors := make([][]int, len(rows))
	for i := range rows {
		for j := range rows {
			if !p.present[i] || !p.present[j] {
				continue
			}
			isNear := i == j
			if !isNear {
				if isNear, err = p.withinDistance(i, j, eps); err != nil {
					return nil, err
				}
			}
			if isNear {
				neighbors[i] = append(neighbors[i], j)
			}
		}
	}
	isCore := func(i int) bool {
		return len(neighbors[i]) >= int(minPoints)
	}
	u := newUnionFind(len(rows))
	clustered := make([]bool, len(rows))
	for i := range rows {
		if !isCore(i) {
			continue
		}
		clustered[i] = true
		for _, j := range neighbors[i] {
			if isCore(j) {
				u.union(i, j)
			} else if !clustered[j] {
				clustered[j] = true
				u[j] = u.find(i)
			}
		}
	}
	return clusterNumbers(u, func(i int) bool { return !clustered[i] }, func(int) bool { return false }), nil
}

// kmeansInput is a geometry that ST_ClusterKMeans clusters as a weighted point.
type kmeansInput struct {
	row    int
	point  [3]float64
	weight float64
}

// newKMeansInput returns the weighted point that stands in for a geometry that is not empty. A point is itself, with
// its M value as its weight. Every other geometry has a weight of one and stands in as the center of its box when it
// has Z values, or as its centroid otherwise.
func newKMeansInput(g *spatial.GEOS, row int, s shape) (kmeansInput, error) {
	b, _ := s.bounds()
	input := kmeansInput{row: row, point: [3]float64{b.minX, b.minY, b.minZ}, weight: 1}
	if s.typeNum == pointType {
		if s.hasM() {
			if b.minM <= 0 {
				return kmeansInput{}, pgerror.New(pgcode.Internal, "lwgeom_cluster_kmeans has an input point geometry with weight in M less or equal to 0")
			}
			input.weight = b.minM
		}
		return input, nil
	}
	if s.hasZ() {
		input.point = [3]float64{(b.minX + b.maxX) / 2, (b.minY + b.maxY) / 2, (b.minZ + b.maxZ) / 2}
		return input, nil
	}
	geosS, err := s.toGEOS()
	if err != nil {
		return kmeansInput{}, err
	}
	centroid, err := g.Centroid(geosS)
	if err != nil {
		return kmeansInput{}, geosError("GEOSGetCentroid", err)
	}
	input.point = [3]float64{centroid.Coords[0], centroid.Coords[1], 0}
	return input, nil
}

// squaredDistance returns the squared distance between two points.
func squaredDistance(a [3]float64, b [3]float64) float64 {
	x, y, z := a[0]-b[0], a[1]-b[1], a[2]-b[2]
	return x*x + y*y + z*z
}

// kmeans splits the inputs into at most `k` clusters and returns the cluster of each input. A single cluster starts
// at the first input. Otherwise, the first two centers are the two inputs farthest apart, taking the later one first,
// and each further center is the input farthest from its nearest center. Then each input repeatedly joins its nearest
// center, preferring the lower cluster on a tie, and each center moves to the weighted mean of its inputs, until no
// input changes cluster.
func kmeans(inputs []kmeansInput, k int) []int {
	var seeds []int
	if len(inputs) == 1 || k == 1 {
		seeds = []int{0}
	} else {
		first, second, farthest := 1, 0, -1.0
		for i := range inputs {
			for j := i + 1; j < len(inputs); j++ {
				if d := squaredDistance(inputs[i].point, inputs[j].point); d > farthest {
					first, second, farthest = j, i, d
				}
			}
		}
		seeds = []int{first, second}
	}
	for len(seeds) < k && len(seeds) < len(inputs) {
		next, nextDistance := -1, -1.0
		for i, input := range inputs {
			nearest := math.Inf(1)
			for _, seed := range seeds {
				if seed == i {
					nearest = -1
					break
				}
				nearest = math.Min(nearest, squaredDistance(input.point, inputs[seed].point))
			}
			if nearest > nextDistance {
				next, nextDistance = i, nearest
			}
		}
		seeds = append(seeds, next)
	}
	centers := make([][3]float64, len(seeds))
	for c, seed := range seeds {
		centers[c] = inputs[seed].point
	}
	clusters := make([]int, len(inputs))
	for i := range clusters {
		clusters[i] = -1
	}
	for iteration := 0; iteration < 1000; iteration++ {
		changed := false
		for i, input := range inputs {
			nearest := 0
			for c := range centers {
				if squaredDistance(input.point, centers[c]) < squaredDistance(input.point, centers[nearest]) {
					nearest = c
				}
			}
			if clusters[i] != nearest {
				clusters[i], changed = nearest, true
			}
		}
		if !changed {
			break
		}
		sums, weights := make([][3]float64, len(centers)), make([]float64, len(centers))
		for i, input := range inputs {
			for d := range sums[clusters[i]] {
				sums[clusters[i]][d] += input.point[d] * input.weight
			}
			weights[clusters[i]] += input.weight
		}
		for c := range centers {
			if weights[c] > 0 {
				centers[c] = [3]float64{sums[c][0] / weights[c], sums[c][1] / weights[c], sums[c][2] / weights[c]}
			}
		}
	}
	return clusters
}

// kmeansRadius returns the distance from the weighted mean of the inputs to the farthest input.
func kmeansRadius(inputs []kmeansInput) float64 {
	var center [3]float64
	var weight float64
	for _, input := range inputs {
		for d := range center {
			center[d] += input.point[d] * input.weight
		}
		weight += input.weight
	}
	for d := range center {
		center[d] /= weight
	}
	radius := 0.0
	for _, input := range inputs {
		radius = math.Max(radius, squaredDistance(input.point, center))
	}
	return math.Sqrt(radius)
}

// clusterKMeans implements ST_ClusterKMeans, which numbers each row by its cluster among `k` clusters, or -1 for a row
// without a geometry. With a positive `max_radius`, each cluster that reaches farther than it from its center is then
// split in two, with the half of the first center keeping the cluster's number and the other half taking the next
// unused number, until every cluster fits.
func clusterKMeans(ctx *sql.Context, rows [][]any) ([]any, error) {
	results := make([]any, len(rows))
	if len(rows) == 0 {
		return results, nil
	}
	k, ok := rows[0][1].(int32)
	if !ok || k <= 0 {
		return results, nil
	}
	if int(k) > len(rows) {
		return nil, pgerror.Newf(pgcode.Internal, "K (%d) must be smaller than the number of rows in the group (%d)", k, len(rows))
	}
	maxRadius, _ := rows[0][2].(float64)
	g := spatial.NewGEOS()
	defer g.Close()
	var inputs []kmeansInput
	distinct := make(map[[3]float64]bool)
	for i, row := range rows {
		results[i] = int32(-1)
		geometry, ok := row[0].(Geometry)
		if !ok || geometry.Shape.isEmpty() {
			continue
		}
		input, err := newKMeansInput(g, i, geometry.Shape)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
		distinct[input.point] = true
	}
	if len(inputs) < int(k) {
		sendNotice(ctx, "NOTICE", fmt.Sprintf("lwgeom_cluster_kmeans: number of non-empty geometries (%d) is less than the number of clusters (%d) requested, not all clusters will get data", len(inputs), k))
	}
	if k > 1 && len(distinct) < len(inputs) {
		sendNotice(ctx, "NOTICE", fmt.Sprintf("kmeans_init: there are at least %d duplicate inputs, number of output clusters may be less than you requested", len(inputs)-len(distinct)+1))
	}
	clusters := kmeans(inputs, int(k))
	if maxRadius > 0 {
		count := int(k)
		for c := 0; c < count; c++ {
			for {
				var members []kmeansInput
				var memberIndexes []int
				for i, cluster := range clusters {
					if cluster == c {
						members = append(members, inputs[i])
						memberIndexes = append(memberIndexes, i)
					}
				}
				if len(members) < 2 || kmeansRadius(members) <= maxRadius {
					break
				}
				for i, half := range kmeans(members, 2) {
					if half == 1 {
						clusters[memberIndexes[i]] = count
					}
				}
				count++
			}
		}
	}
	for i, input := range inputs {
		results[input.row] = int32(clusters[i])
	}
	return results, nil
}

// clusterRoutines returns the window functions that cluster geometries.
func clusterRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_clusterdbscan", Symbol: "ST_ClusterDBSCAN", Parameters: []extdef.Parameter{{Type: "geometry"}, {Name: "eps", Type: "float8"}, {Name: "minpoints", Type: "int4"}}, Returns: "int4", Strict: true, WindowImpl: clusterDBSCAN},
		{Name: "st_clusterintersectingwin", Symbol: "ST_ClusterIntersectingWin", Parameters: params("geometry"), Returns: "int4", Strict: true, WindowImpl: clusterWindow(func(p *partitionGeometries, i int, j int, row []any) (bool, error) {
			return p.intersects(i, j)
		})},
		{Name: "st_clusterkmeans", Symbol: "ST_ClusterKMeans", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "k", Type: "int4"}, {Name: "max_radius", Type: "float8", Default: "NULL::double precision"}}, Returns: "int4", Strict: true, WindowImpl: clusterKMeans},
		{Name: "st_clusterwithinwin", Symbol: "ST_ClusterWithinWin", Parameters: []extdef.Parameter{{Type: "geometry"}, {Name: "distance", Type: "float8"}}, Returns: "int4", Strict: true, WindowImpl: clusterWindow(func(p *partitionGeometries, i int, j int, row []any) (bool, error) {
			distance, ok := row[1].(float64)
			if !ok {
				return false, nil
			}
			return p.withinDistance(i, j, distance)
		})},
	}
}

// geosTooOld returns an implementation of a coverage window function that needs a newer version of GEOS than the one
// that the extension links, which always fails as PostGIS does when it is built against that GEOS.
func geosTooOld(message string) extdef.WindowFunction {
	return func(ctx *sql.Context, rows [][]any) ([]any, error) {
		//TODO: implement once the extension links GEOS 3.14 or later
		return nil, pgerror.New(pgcode.Internal, message)
	}
}

// coverageRoutines returns the window functions that check, simplify, and clean polygonal coverages.
func coverageRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_coverageclean", Symbol: "ST_CoverageClean", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "gapmaximumwidth", Type: "float8", Default: "0.0"}, {Name: "snappingdistance", Type: "float8", Default: "'-1.0'::numeric"}, {Name: "overlapmergestrategy", Type: "text", Default: "'MERGE_LONGEST_BORDER'::text"}}, Returns: "geometry", Strict: true,
			WindowImpl: geosTooOld("The GEOS version this PostGIS binary was compiled against (31101) not include 'GEOSCoverageClean' function (3.14 or greater required)")},
		{Name: "st_coverageinvalidedges", Symbol: "ST_CoverageInvalidEdges", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "tolerance", Type: "float8", Default: "0.0"}}, Returns: "geometry", Strict: true,
			WindowImpl: geosTooOld("The GEOS version this PostGIS binary was compiled against (31101) does not include 'GEOSCoverageIsValid' function (3.12 or greater required)")},
		{Name: "st_coveragesimplify", Symbol: "ST_CoverageSimplify", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "tolerance", Type: "float8"}, {Name: "simplifyboundary", Type: "bool", Default: "true"}}, Returns: "geometry", Strict: true,
			WindowImpl: geosTooOld("The GEOS version this PostGIS binary was compiled against (31101) not include 'GEOSCoverageSimplifyVW' function (3.12 or greater required)")},
	}
}
