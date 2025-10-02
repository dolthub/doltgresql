// Copyright 2025 Dolthub, Inc.
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

package _go

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/testing/dumps"
)

// TestImportingDumps are regression tests against dumps taken from various sources.
func TestImportingDumps(t *testing.T) {
	t.Skip("The majority fail for now")
	RunImportTests(t, []ImportTest{
		{
			Name: "Scrubbed-1",
			SetUpScript: []string{
				"CREATE USER behfjgnf WITH SUPERUSER PASSWORD 'password';",
			},
			SkipQueries: []string{"CREATE UNIQUE INDEX dawkmezfehakyikllr"},
			SQLFilename: "scrubbed-1.sql",
		},
		{
			Name:        "A-lang209/Salon-Appointment-Scheduler",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "A-lang209_Salon-Appointment-Scheduler.sql",
		},
		{
			Name:        "Abhishek842000/DB-Performance-Comparator",
			SQLFilename: "Abhishek842000_DB-Performance-Comparator.sql",
		},
		{
			Name:        "AlexTransit/venderctl",
			SQLFilename: "AlexTransit_venderctl.sql",
		},
		{
			Name:        "AliiAhmadi/PostScan",
			SQLFilename: "AliiAhmadi_PostScan.sql",
		},
		{
			Name:        "amittannagit/World-Cup-database-project-files",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "amittannagit_World-Cup-database-project-files.sql",
		},
		{
			Name:        "Ansh-Rathod/Musive-backend-2.0",
			SQLFilename: "Ansh-Rathod_Musive-backend-2.0.sql",
		},
		{
			Name:        "artygg/Data-Processing-Goida",
			SQLFilename: "artygg_Data-Processing-Goida.sql",
		},
		{
			Name:        "bartr/agency",
			SQLFilename: "bartr_agency.sql",
		},
		{
			Name:        "bclynch/edmflare",
			SQLFilename: "bclynch_edmflare.sql",
		},
		{
			Name:        "Billoxinogen18/ar_backend",
			SQLFilename: "Billoxinogen18_ar_backend.sql",
		},
		{
			Name:        "blacktscoder/CrisisSolver",
			SQLFilename: "blacktscoder_CrisisSolver.sql",
		},
		{
			Name:        "Boluwatife-AJB/backend-in-node",
			SQLFilename: "Boluwatife-AJB_backend-in-node.sql",
		},
		{
			Name:        "bonf1re/campis",
			SQLFilename: "bonf1re_campis.sql",
		},
		{
			Name:        "by-zah/universityScheduleBot",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "by-zah_universityScheduleBot.sql",
		},
		{
			Name:        "cardox6/pagila",
			SQLFilename: "cardox6_pagila.sql",
		},
		{
			Name:        "Chris-Merced/Classic-Messenger-App-Backend",
			SQLFilename: "Chris-Merced_Classic-Messenger-App-Backend.sql",
		},
		{
			Name:        "cipherstash/pyconau2024-ctf",
			SQLFilename: "cipherstash_pyconau2024-ctf.sql",
		},
		{
			Name:        "Clar17y/Football-Events",
			SQLFilename: "Clar17y_Football-Events.sql",
		},
		{
			Name:        "CollegeFootballRisk/Risk",
			SQLFilename: "CollegeFootballRisk_Risk.sql",
		},
		{
			Name:        "CONABIO/inaturalist_snmb",
			SQLFilename: "CONABIO_inaturalist_snmb.sql",
		},
		{
			Name:        "CravingCrates/AnkiCollab-Backend",
			SQLFilename: "CravingCrates_AnkiCollab-Backend.sql",
		},
		{
			Name:        "cskerritt/lifeplan-genius",
			SQLFilename: "cskerritt_lifeplan-genius.sql",
		},
		{
			Name:        "dbarrera98/proyecto-informa",
			SQLFilename: "dbarrera98_proyecto-informa.sql",
		},
		{
			Name:        "dennis-campos-11/xg90_app",
			SQLFilename: "dennis-campos-11_xg90_app.sql",
		},
		{
			Name:        "DmitryAntipin151002/Diplom",
			SQLFilename: "DmitryAntipin151002_Diplom.sql",
		},
		{
			Name:        "Dmitrytsg/onectest",
			SQLFilename: "Dmitrytsg_onectest.sql",
		},
		{
			Name:        "DRON12261/EduVault",
			SQLFilename: "DRON12261_EduVault.sql",
		},
		{
			Name:        "DTOcean/dtocean-database",
			SQLFilename: "DTOcean_dtocean-database.sql",
		},
		{
			Name:        "EdwinRo121/ParcialApi",
			SQLFilename: "EdwinRo121_ParcialApi.sql",
		},
		{
			Name:        "Enesuygurs/steamcafe",
			SQLFilename: "Enesuygurs_steamcafe.sql",
		},
		{
			Name:        "erlitx/sql_final",
			SQLFilename: "erlitx_sql_final.sql",
		},
		{
			Name:        "ExposedCat/cashiers-in-shop",
			SQLFilename: "ExposedCat_cashiers-in-shop.sql",
		},
		{
			Name:        "falling-fruit/falling-fruit",
			SQLFilename: "falling-fruit_falling-fruit.sql",
		},
		{
			Name:        "fanfanfw/bdt_rest-api-scraping-result",
			SQLFilename: "fanfanfw_bdt_rest-api-scraping-result.sql",
		},
		{
			Name:        "fn-bucket/fnb-nuxt-postgraphile",
			SQLFilename: "fn-bucket_fnb-nuxt-postgraphile.sql",
		},
		{
			Name:        "Freeztyle17/Neoflex_1",
			SQLFilename: "Freeztyle17_Neoflex_1.sql",
		},
		{
			Name:        "gabrundo/Progetto-Basi-Dati",
			SQLFilename: "gabrundo_Progetto-Basi-Dati.sql",
		},
		{
			Name:        "gnsnghm/cms",
			SQLFilename: "gnsnghm_cms.sql",
		},
		{
			Name:        "gsdnMartin/PIDAP",
			SQLFilename: "gsdnMartin_PIDAP.sql",
		},
		{
			Name:        "HalfCoke/blog_img",
			SQLFilename: "HalfCoke_blog_img.sql",
		},
		{
			Name:        "HarukaMa/aleo-explorer",
			SQLFilename: "HarukaMa_aleo-explorer.sql",
		},
		{
			Name:        "heydabop/rustyz",
			SQLFilename: "heydabop_rustyz.sql",
		},
		{
			Name:        "HugoTZC/OASA",
			SQLFilename: "HugoTZC_OASA.sql",
		},
		{
			Name:        "iangow/pg_functions",
			SQLFilename: "iangow_pg_functions.sql",
		},
		{
			Name:        "ii-habibi/Dental-Clinic",
			SQLFilename: "ii-habibi_Dental-Clinic.sql",
		},
		{
			Name:        "ilovejs/Go-Echo-Boiler",
			SQLFilename: "ilovejs_Go-Echo-Boiler.sql",
		},
		{
			Name:        "InnovaTech-Official/LMS",
			SQLFilename: "InnovaTech-Official_LMS.sql",
		},
		{
			Name:        "jeffchang001/ee-midd",
			SQLFilename: "jeffchang001_ee-midd.sql",
		},
		{
			Name:        "joaoporto27/Bora-Viajar-BackEnd",
			SQLFilename: "joaoporto27_Bora-Viajar-BackEnd.sql",
		},
		{
			Name:        "joec05/social-media-app-pgsql",
			SQLFilename: "joec05_social-media-app-pgsql.sql",
		},
		{
			Name:        "julesd7/collatask",
			SQLFilename: "julesd7_collatask.sql",
		},
		{
			Name:        "jwalit21/BitmapJoinDatabaseEngine",
			SQLFilename: "jwalit21_BitmapJoinDatabaseEngine.sql",
		},
		{
			Name:        "KangAbbad/laundry-app",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "KangAbbad_laundry-app.sql",
		},
		{
			Name:        "kapil23jani/hospitease_backend",
			SQLFilename: "kapil23jani_hospitease_backend.sql",
		},
		{
			Name:        "kentyler/conversationalaiapi",
			SQLFilename: "kentyler_conversationalaiapi.sql",
		},
		{
			Name:        "kepinskw/db-jobportal",
			SQLFilename: "kepinskw_db-jobportal.sql",
		},
		{
			Name:        "kirooha/adtech-simple",
			SQLFilename: "kirooha_adtech-simple.sql",
		},
		{
			Name:        "kjanus03/tsn",
			SQLFilename: "kjanus03_tsn.sql",
		},
		{
			Name:        "kraftn/queue-server",
			SQLFilename: "kraftn_queue-server.sql",
		},
		{
			Name:        "linvivian7/fe-react-16-demo",
			SQLFilename: "linvivian7_fe-react-16-demo.sql",
		},
		{
			Name:        "littlebunch/graphql-rs",
			SQLFilename: "littlebunch_graphql-rs.sql",
		},
		{
			Name:        "luizantoniocardoso/trabalho-banco-2",
			SQLFilename: "luizantoniocardoso_trabalho-banco-2.sql",
		},
		{
			Name:        "mintas123/Buddies-API",
			SQLFilename: "mintas123_Buddies-API.sql",
		},
		{
			Name:        "Mistral-war2ru/PG-connect",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "Mistral-war2ru_PG-connect.sql",
		},
		{
			Name:        "mostafacs/ecommerce-microservices-spring-reactive-webflux",
			SQLFilename: "mostafacs_ecommerce-microservices-spring-reactive-webflux.sql",
		},
		{
			Name:        "MostafaProgramming/100719549",
			SQLFilename: "MostafaProgramming_100719549.sql",
		},
		{
			Name:        "mraescudeiro/subclue",
			SQLFilename: "mraescudeiro_subclue.sql",
		},
		{
			Name:        "mvnp/start-dashboard-v3-backend",
			SQLFilename: "mvnp_start-dashboard-v3-backend.sql",
		},
		{
			Name:        "NarasimhaProcess/UserTracking",
			SQLFilename: "NarasimhaProcess_UserTracking.sql",
		},
		{
			Name:        "NathalyHolguin16/Sistema_de_gesti-n_de_Cine",
			SQLFilename: "NathalyHolguin16_Sistema_de_gesti-n_de_Cine.sql",
		},
		{
			Name:        "NECKER55/supermarket_shop",
			SQLFilename: "NECKER55_supermarket_shop.sql",
		},
		{
			Name:        "nxtrm/neanote",
			SQLFilename: "nxtrm_neanote.sql",
		},
		{
			Name:        "nyfagel/klubb",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "nyfagel_klubb.sql",
		},
		{
			Name:        "oknosoft/windowbuilder-planning",
			SQLFilename: "oknosoft_windowbuilder-planning.sql",
		},
		{
			Name:        "openeventdatabase/backend",
			SQLFilename: "openeventdatabase_backend.sql",
		},
		{
			Name:        "openlawnz/openlawnz-data-processor",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "openlawnz_openlawnz-data-processor.sql",
		},
		{
			Name:        "oslabs-beta/ditto",
			SQLFilename: "oslabs-beta_ditto.sql",
		},
		{
			Name:        "paulshriner/fcc-rd-cert",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "paulshriner_fcc-rd-cert.sql",
		},
		{
			Name:        "qqtati/diplom",
			SQLFilename: "qqtati_diplom.sql",
		},
		{
			Name:        "riclolsen/json-scada",
			SQLFilename: "riclolsen_json-scada.sql",
		},
		{
			Name:        "rmarquez123/titans",
			SQLFilename: "rmarquez123_titans.sql",
		},
		{
			Name:        "roboflow/scavenger-hunt",
			SQLFilename: "roboflow_scavenger-hunt.sql",
		},
		{
			Name:        "RSNA/isn-edge-server-database",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "RSNA_isn-edge-server-database.sql",
		},
		{
			Name:        "S0mbre/russtat",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "S0mbre_russtat.sql",
		},
		{
			Name:        "slsfi/digital_edition_db",
			SQLFilename: "slsfi_digital_edition_db.sql",
		},
		{
			Name:        "SoniaMarrocco/Library-Database",
			SQLFilename: "SoniaMarrocco_Library-Database.sql",
		},
		{
			Name:        "StrangeGoofy/Yota_game",
			SQLFilename: "StrangeGoofy_Yota_game.sql",
		},
		{
			Name:        "StronglogicSolutions/kserver",
			SQLFilename: "StronglogicSolutions_kserver.sql",
		},
		{
			Name:        "surgefm/v2land-redstone",
			SQLFilename: "surgefm_v2land-redstone.sql",
		},
		{
			Name:        "sylvain-guehria/StockShop",
			SQLFilename: "sylvain-guehria_StockShop.sql",
		},
		{
			Name:        "TaraPadilla/MarketSpring",
			SQLFilename: "TaraPadilla_MarketSpring.sql",
		},
		{
			Name:        "the-benchmarker/web-frameworks",
			SQLFilename: "the-benchmarker_web-frameworks.sql",
		},
		{
			Name:        "theophoric/prisma-near-indexer",
			SQLFilename: "theophoric_prisma-near-indexer.sql",
		},
		{
			Name:        "Timovski/Co-opMinesweeper",
			SQLFilename: "Timovski_Co-opMinesweeper.sql",
		},
		{
			Name:        "tpolecat/cofree",
			SQLFilename: "tpolecat_cofree.sql",
		},
		{
			Name:        "uzurpastor/UniWorks",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "uzurpastor_UniWorks.sql",
		},
		{
			Name:        "Vanesor/zenith",
			SQLFilename: "Vanesor_zenith.sql",
		},
		{
			Name:        "Vinicius02612/sistema_da_associacao",
			SQLFilename: "Vinicius02612_sistema_da_associacao.sql",
		},
		{
			Name:        "WhoIsKatie/e-Hotels",
			Skip:        true, // Database creation uses unsupported params then attempts to connect, hangs indefinitely
			SQLFilename: "WhoIsKatie_e-Hotels.sql",
		},
		{
			Name:        "wolfufu/Hakaton2025Spring",
			SQLFilename: "wolfufu_Hakaton2025Spring.sql",
		},
		{
			Name:        "Xarpunk/DemExam",
			SQLFilename: "Xarpunk_DemExam.sql",
		},
		{
			Name:        "yase-search/yase-engine",
			SQLFilename: "yase-search_yase-engine.sql",
		},
		{
			Name:        "Yesk0/KBTU_Database_24-25",
			SQLFilename: "Yesk0_KBTU_Database_24-25.sql",
		},
	})
}

// TriggerImportBreakpoint exists so that a breakpoint may be set within the function, on the unused Sprintf. This
// function is called whenever a query matches one of the breakpoint queries defined in the import test. This enables us
// to simulate some kind of breakpoint functionality on import queries, which isn't normally possible.
func TriggerImportBreakpoint(breakpointQuery string) {
	// It doesn't actually matter what this function is. It's just here so we can set a breakpoint on something.
	_ = fmt.Sprintf("__%s", breakpointQuery)
}

// ImportTest is a test for importing SQL dumps.
type ImportTest struct {
	Name        string
	SetUpScript []string
	Focus       bool
	Skip        bool
	SQLFilename string
	// Breakpoints allow for triggering breakpoints when any matching queries are given. A breakpoint must be set within
	// TriggerImportBreakpoint for this to work.
	Breakpoints []string
	// SkipQueries
	SkipQueries []string
}

// RunImportTests runs the given ImportTest scripts.
func RunImportTests(t *testing.T, scripts []ImportTest) {
	if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
		if _, ok = os.LookupEnv("GITHUB_ACTION_IMPORT_DUMPS"); !ok {
			t.Skip("These tests are run in their own dedicated action")
		}
	}
	var psqlCommand string
	switch runtime.GOOS {
	case "windows":
		psqlCommand = "psql.exe"
	default:
		psqlCommand = "psql"
	}
	// Check if PSQL runs directly
	var outBuffer bytes.Buffer
	cmd := exec.Command(psqlCommand, "--version")
	cmd.Stdout = &outBuffer
	if !assert.NoError(t, cmd.Run()) || !strings.Contains(outBuffer.String(), "PostgreSQL") {
		// We could not run PSQL and get the version, so it must not be in the path.
		// We'll check if pg_config is in the path and reference the binary directly.
		outBuffer.Reset()
		cmd = exec.Command("pg_config", "--bindir")
		cmd.Stdout = &outBuffer
		if !assert.NoError(t, cmd.Run()) {
			require.Fail(t, "Postgres is not installed, cannot run tests")
		}
		psqlCommand = filepath.Join(strings.TrimSpace(outBuffer.String()), psqlCommand)
		// pg_config is in the path, so we'll try and run PSQL by directly referencing the binary
		outBuffer.Reset()
		cmd = exec.Command(psqlCommand, "--version")
		cmd.Stdout = &outBuffer
		if !assert.NoError(t, cmd.Run()) || !strings.Contains(outBuffer.String(), "PostgreSQL") {
			t.Fatalf("PSQL cannot be found at: `%s`", psqlCommand)
		}
	}
	// Grab the folder with the files to import
	_, currentFileLocation, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Unable to find the folder where the files to import are located")
	}
	dumpsFolder := filepath.Clean(filepath.Join(filepath.Dir(currentFileLocation), "../dumps/"))
	// Set whether we're checking Focus-only scripts or not
	useFocus := false
	for _, script := range scripts {
		if script.Focus {
			// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
			if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
				panic(fmt.Sprintf("The script `%s` has Focus set to `true`. GitHub Actions requires that "+
					"all tests are run, which Focus circumvents, leading to this error. Please disable Focus on "+
					"all tests.", script.Name))
			}
			useFocus = true
			break
		}
	}
	for _, script := range scripts {
		if useFocus != script.Focus {
			continue
		}
		RunImportTest(t, script, psqlCommand, dumpsFolder)
	}
}

// RunImportTest runs the given ImportTest script.
func RunImportTest(t *testing.T, script ImportTest, psqlCommand string, dumpsFolder string) {
	// TODO: handle other dump types, such as those that require pg_restore
	t.Run(script.Name, func(t *testing.T) {
		// Mark this test as skipped if we have it set
		if script.Skip {
			t.Skip()
		}
		// Create the in-memory server that we'll test against
		port, err := sql.GetEmptyPort()
		require.NoError(t, err)
		ctx, conn, controller := CreateServerWithPort(t, "postgres", port)
		func() {
			defer conn.Close(ctx)
			for _, query := range script.SetUpScript {
				_, err = conn.Exec(ctx, query)
				require.NoError(t, err)
			}
		}()
		defer func() {
			controller.Stop()
			err := controller.WaitForStop()
			require.NoError(t, err)
		}()
		// Create the message interceptor
		var qeChan chan dumps.ImportQueryError
		port, qeChan = dumps.InterceptImportMessages(t, dumps.InterceptArgs{
			DoltgresPort:      port,
			SkippedQueries:    script.SkipQueries,
			BreakpointQueries: script.Breakpoints,
			TriggerBreakpoint: TriggerImportBreakpoint,
		})
		defer close(qeChan)
		var allErrors []dumps.ImportQueryError
		go func() {
			for chanErr := range qeChan {
				allErrors = append(allErrors, chanErr)
			}
		}()
		// Run the import
		var outBuffer bytes.Buffer
		var errBuffer bytes.Buffer
		cmd := exec.Command(psqlCommand, fmt.Sprintf("postgresql://postgres:password@localhost:%d/postgres?sslmode=disable", port))
		cmd.Stdout = &outBuffer
		cmd.Stderr = &errBuffer
		targetFile, err := os.Open(filepath.Join(dumpsFolder, "sql", script.SQLFilename))
		require.NoError(t, err)
		cmd.Stdin = targetFile
		require.NoError(t, cmd.Run())
		if len(allErrors) > 0 {
			t.Logf("COUNT: %d", len(allErrors))
			// If we have more than some threshold, then we'll only show the first few for ease of consumption
			for i := 0; i < len(allErrors) && i < 10; i++ {
				t.Logf("QUERY: %s\nERROR: %s", allErrors[i].Query, allErrors[i].Error)
			}
			t.FailNow()
		}
	})
}
