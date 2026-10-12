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

use harness::dumps::{ImportTest, run_import_test};

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn scrubbed_1() {
    run_import_test(&ImportTest {
        name: "Scrubbed-1",
        set_up_script: &["CREATE USER behfjgnf WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "scrubbed-1.sql",
        skip_queries: &["CREATE UNIQUE INDEX dawkmezfehakyikllr"],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn a_lang209_salon_appointment_scheduler() {
    run_import_test(&ImportTest {
        name: "A-lang209/Salon-Appointment-Scheduler",
        set_up_script: &[],
        sql_filename: "A-lang209_Salon-Appointment-Scheduler.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn abhishek842000_db_performance_comparator() {
    run_import_test(&ImportTest {
        name: "Abhishek842000/DB-Performance-Comparator",
        set_up_script: &[],
        sql_filename: "Abhishek842000_DB-Performance-Comparator.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn alextransit_venderctl() {
    run_import_test(&ImportTest {
        name: "AlexTransit/venderctl",
        set_up_script: &[],
        sql_filename: "AlexTransit_venderctl.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn aliiahmadi_postscan() {
    run_import_test(&ImportTest {
        name: "AliiAhmadi/PostScan",
        set_up_script: &["CREATE USER testuser WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "AliiAhmadi_PostScan.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn amittannagit_world_cup_database_project_files() {
    run_import_test(&ImportTest {
        name: "amittannagit/World-Cup-database-project-files",
        set_up_script: &["CREATE USER freecodecamp WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "amittannagit_World-Cup-database-project-files.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn ansh_rathod_musive_backend_2_0() {
    run_import_test(&ImportTest {
        name: "Ansh-Rathod/Musive-backend-2.0",
        set_up_script: &[],
        sql_filename: "Ansh-Rathod_Musive-backend-2.0.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn artygg_data_processing_goida() {
    run_import_test(&ImportTest {
        name: "artygg/Data-Processing-Goida",
        set_up_script: &[
            "CREATE USER app_admin WITH SUPERUSER PASSWORD 'password';",
            "CREATE USER analytics_viewer WITH SUPERUSER PASSWORD 'password';",
            "CREATE USER content_manager WITH SUPERUSER PASSWORD 'password';",
        ],
        sql_filename: "artygg_Data-Processing-Goida.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn bartr_agency() {
    run_import_test(&ImportTest {
        name: "bartr/agency",
        set_up_script: &[],
        sql_filename: "bartr_agency.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn bclynch_edmflare() {
    run_import_test(&ImportTest {
        name: "bclynch/edmflare",
        set_up_script: &["CREATE USER edm WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "bclynch_edmflare.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn billoxinogen18_ar_backend() {
    run_import_test(&ImportTest {
        name: "Billoxinogen18/ar_backend",
        set_up_script: &[],
        sql_filename: "Billoxinogen18_ar_backend.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn blacktscoder_crisissolver() {
    run_import_test(&ImportTest {
        name: "blacktscoder/CrisisSolver",
        set_up_script: &["CREATE USER crisisresolver_visitor WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "blacktscoder_CrisisSolver.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn boluwatife_ajb_backend_in_node() {
    run_import_test(&ImportTest {
        name: "Boluwatife-AJB/backend-in-node",
        set_up_script: &[],
        sql_filename: "Boluwatife-AJB_backend-in-node.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn bonf1re_campis() {
    run_import_test(&ImportTest {
        name: "bonf1re/campis",
        set_up_script: &[],
        sql_filename: "bonf1re_campis.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn by_zah_universityschedulebot() {
    run_import_test(&ImportTest {
        name: "by-zah/universityScheduleBot",
        set_up_script: &[],
        sql_filename: "by-zah_universityScheduleBot.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn cardox6_pagila() {
    run_import_test(&ImportTest {
        name: "cardox6/pagila",
        set_up_script: &[],
        sql_filename: "cardox6_pagila.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn chris_merced_classic_messenger_app_backend() {
    run_import_test(&ImportTest {
        name: "Chris-Merced/Classic-Messenger-App-Backend",
        set_up_script: &[],
        sql_filename: "Chris-Merced_Classic-Messenger-App-Backend.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn cipherstash_pyconau2024_ctf() {
    run_import_test(&ImportTest {
        name: "cipherstash/pyconau2024-ctf",
        set_up_script: &[],
        sql_filename: "cipherstash_pyconau2024-ctf.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn clar17y_football_events() {
    run_import_test(&ImportTest {
        name: "Clar17y/Football-Events",
        set_up_script: &[],
        sql_filename: "Clar17y_Football-Events.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn collegefootballrisk_risk() {
    run_import_test(&ImportTest {
        name: "CollegeFootballRisk/Risk",
        set_up_script: &["CREATE USER risk WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "CollegeFootballRisk_Risk.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn conabio_inaturalist_snmb() {
    run_import_test(&ImportTest {
        name: "CONABIO/inaturalist_snmb",
        set_up_script: &[],
        sql_filename: "CONABIO_inaturalist_snmb.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn cravingcrates_ankicollab_backend() {
    run_import_test(&ImportTest {
        name: "CravingCrates/AnkiCollab-Backend",
        set_up_script: &[],
        sql_filename: "CravingCrates_AnkiCollab-Backend.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn cskerritt_lifeplan_genius() {
    run_import_test(&ImportTest {
        name: "cskerritt/lifeplan-genius",
        set_up_script: &[],
        sql_filename: "cskerritt_lifeplan-genius.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn dbarrera98_proyecto_informa() {
    run_import_test(&ImportTest {
        name: "dbarrera98/proyecto-informa",
        set_up_script: &[],
        sql_filename: "dbarrera98_proyecto-informa.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn dennis_campos_11_xg90_app() {
    run_import_test(&ImportTest {
        name: "dennis-campos-11/xg90_app",
        set_up_script: &[],
        sql_filename: "dennis-campos-11_xg90_app.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn dmitryantipin151002_diplom() {
    run_import_test(&ImportTest {
        name: "DmitryAntipin151002/Diplom",
        set_up_script: &[],
        sql_filename: "DmitryAntipin151002_Diplom.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn dmitrytsg_onectest() {
    run_import_test(&ImportTest {
        name: "Dmitrytsg/onectest",
        set_up_script: &[],
        sql_filename: "Dmitrytsg_onectest.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn dron12261_eduvault() {
    run_import_test(&ImportTest {
        name: "DRON12261/EduVault",
        set_up_script: &[],
        sql_filename: "DRON12261_EduVault.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn dtocean_dtocean_database() {
    run_import_test(&ImportTest {
        name: "DTOcean/dtocean-database",
        set_up_script: &[],
        sql_filename: "DTOcean_dtocean-database.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn edwinro121_parcialapi() {
    run_import_test(&ImportTest {
        name: "EdwinRo121/ParcialApi",
        set_up_script: &[],
        sql_filename: "EdwinRo121_ParcialApi.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn enesuygurs_steamcafe() {
    run_import_test(&ImportTest {
        name: "Enesuygurs/steamcafe",
        set_up_script: &[],
        sql_filename: "Enesuygurs_steamcafe.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn erlitx_sql_final() {
    run_import_test(&ImportTest {
        name: "erlitx/sql_final",
        set_up_script: &[],
        sql_filename: "erlitx_sql_final.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn exposedcat_cashiers_in_shop() {
    run_import_test(&ImportTest {
        name: "ExposedCat/cashiers-in-shop",
        set_up_script: &[r#"CREATE USER "shop-admin" WITH SUPERUSER PASSWORD 'password';"#],
        sql_filename: "ExposedCat_cashiers-in-shop.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn falling_fruit_falling_fruit() {
    run_import_test(&ImportTest {
        name: "falling-fruit/falling-fruit",
        set_up_script: &[],
        sql_filename: "falling-fruit_falling-fruit.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn fanfanfw_bdt_rest_api_scraping_result() {
    run_import_test(&ImportTest {
        name: "fanfanfw/bdt_rest-api-scraping-result",
        set_up_script: &[],
        sql_filename: "fanfanfw_bdt_rest-api-scraping-result.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn fn_bucket_fnb_nuxt_postgraphile() {
    run_import_test(&ImportTest {
        name: "fn-bucket/fnb-nuxt-postgraphile",
        set_up_script: &[],
        sql_filename: "fn-bucket_fnb-nuxt-postgraphile.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn freeztyle17_neoflex_1() {
    run_import_test(&ImportTest {
        name: "Freeztyle17/Neoflex_1",
        set_up_script: &[],
        sql_filename: "Freeztyle17_Neoflex_1.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn gabrundo_progetto_basi_dati() {
    run_import_test(&ImportTest {
        name: "gabrundo/Progetto-Basi-Dati",
        set_up_script: &[],
        sql_filename: "gabrundo_Progetto-Basi-Dati.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn gnsnghm_cms() {
    run_import_test(&ImportTest {
        name: "gnsnghm/cms",
        set_up_script: &[],
        sql_filename: "gnsnghm_cms.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn gsdnmartin_pidap() {
    run_import_test(&ImportTest {
        name: "gsdnMartin/PIDAP",
        set_up_script: &[],
        sql_filename: "gsdnMartin_PIDAP.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn halfcoke_blog_img() {
    run_import_test(&ImportTest {
        name: "HalfCoke/blog_img",
        set_up_script: &["CREATE USER ssouser WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "HalfCoke_blog_img.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn harukama_aleo_explorer() {
    run_import_test(&ImportTest {
        name: "HarukaMa/aleo-explorer",
        set_up_script: &[],
        sql_filename: "HarukaMa_aleo-explorer.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn heydabop_rustyz() {
    run_import_test(&ImportTest {
        name: "heydabop/rustyz",
        set_up_script: &["CREATE USER rustyz WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "heydabop_rustyz.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn hugotzc_oasa() {
    run_import_test(&ImportTest {
        name: "HugoTZC/OASA",
        set_up_script: &[],
        sql_filename: "HugoTZC_OASA.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn iangow_pg_functions() {
    run_import_test(&ImportTest {
        name: "iangow/pg_functions",
        set_up_script: &[],
        sql_filename: "iangow_pg_functions.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn ii_habibi_dental_clinic() {
    run_import_test(&ImportTest {
        name: "ii-habibi/Dental-Clinic",
        set_up_script: &[],
        sql_filename: "ii-habibi_Dental-Clinic.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn innovatech_official_lms() {
    run_import_test(&ImportTest {
        name: "InnovaTech-Official/LMS",
        set_up_script: &[],
        sql_filename: "InnovaTech-Official_LMS.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn jeffchang001_ee_midd() {
    run_import_test(&ImportTest {
        name: "jeffchang001/ee-midd",
        set_up_script: &[],
        sql_filename: "jeffchang001_ee-midd.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn joaoporto27_bora_viajar_backend() {
    run_import_test(&ImportTest {
        name: "joaoporto27/Bora-Viajar-BackEnd",
        set_up_script: &[],
        sql_filename: "joaoporto27_Bora-Viajar-BackEnd.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn joec05_social_media_app_pgsql() {
    run_import_test(&ImportTest {
        name: "joec05/social-media-app-pgsql",
        set_up_script: &[],
        sql_filename: "joec05_social-media-app-pgsql.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn julesd7_collatask() {
    run_import_test(&ImportTest {
        name: "julesd7/collatask",
        set_up_script: &[],
        sql_filename: "julesd7_collatask.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn jwalit21_bitmapjoindatabaseengine() {
    run_import_test(&ImportTest {
        name: "jwalit21/BitmapJoinDatabaseEngine",
        set_up_script: &[],
        sql_filename: "jwalit21_BitmapJoinDatabaseEngine.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn kangabbad_laundry_app() {
    run_import_test(&ImportTest {
        name: "KangAbbad/laundry-app",
        set_up_script: &[],
        sql_filename: "KangAbbad_laundry-app.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn kapil23jani_hospitease_backend() {
    run_import_test(&ImportTest {
        name: "kapil23jani/hospitease_backend",
        set_up_script: &[r#"CREATE USER "hospitease_admin" WITH SUPERUSER PASSWORD 'password';"#],
        sql_filename: "kapil23jani_hospitease_backend.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn kentyler_conversationalaiapi() {
    run_import_test(&ImportTest {
        name: "kentyler/conversationalaiapi",
        set_up_script: &[],
        sql_filename: "kentyler_conversationalaiapi.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn kepinskw_db_jobportal() {
    run_import_test(&ImportTest {
        name: "kepinskw/db-jobportal",
        set_up_script: &["CREATE USER recruiter;", "CREATE USER job_seeker;", "CREATE USER employer;"],
        sql_filename: "kepinskw_db-jobportal.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn kirooha_adtech_simple() {
    run_import_test(&ImportTest {
        name: "kirooha/adtech-simple",
        set_up_script: &[],
        sql_filename: "kirooha_adtech-simple.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn kjanus03_tsn() {
    run_import_test(&ImportTest {
        name: "kjanus03/tsn",
        set_up_script: &[],
        sql_filename: "kjanus03_tsn.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn kraftn_queue_server() {
    run_import_test(&ImportTest {
        name: "kraftn/queue-server",
        set_up_script: &[],
        sql_filename: "kraftn_queue-server.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn linvivian7_fe_react_16_demo() {
    run_import_test(&ImportTest {
        name: "linvivian7/fe-react-16-demo",
        set_up_script: &[],
        sql_filename: "linvivian7_fe-react-16-demo.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn littlebunch_graphql_rs() {
    run_import_test(&ImportTest {
        name: "littlebunch/graphql-rs",
        set_up_script: &[],
        sql_filename: "littlebunch_graphql-rs.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn luizantoniocardoso_trabalho_banco_2() {
    run_import_test(&ImportTest {
        name: "luizantoniocardoso/trabalho-banco-2",
        set_up_script: &[],
        sql_filename: "luizantoniocardoso_trabalho-banco-2.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn mintas123_buddies_api() {
    run_import_test(&ImportTest {
        name: "mintas123/Buddies-API",
        set_up_script: &[],
        sql_filename: "mintas123_Buddies-API.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn mistral_war2ru_pg_connect() {
    run_import_test(&ImportTest {
        name: "Mistral-war2ru/PG-connect",
        set_up_script: &[],
        sql_filename: "Mistral-war2ru_PG-connect.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn mostafacs_ecommerce_microservices_spring_reactive_webflux() {
    run_import_test(&ImportTest {
        name: "mostafacs/ecommerce-microservices-spring-reactive-webflux",
        set_up_script: &[],
        sql_filename: "mostafacs_ecommerce-microservices-spring-reactive-webflux.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn mostafaprogramming_100719549() {
    run_import_test(&ImportTest {
        name: "MostafaProgramming/100719549",
        set_up_script: &[],
        sql_filename: "MostafaProgramming_100719549.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn mraescudeiro_subclue() {
    run_import_test(&ImportTest {
        name: "mraescudeiro/subclue",
        set_up_script: &[],
        sql_filename: "mraescudeiro_subclue.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn mvnp_start_dashboard_v3_backend() {
    run_import_test(&ImportTest {
        name: "mvnp/start-dashboard-v3-backend",
        set_up_script: &["CREATE USER neondb_owner WITH SUPERUSER PASSWORD 'password';"],
        sql_filename: "mvnp_start-dashboard-v3-backend.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn narasimhaprocess_usertracking() {
    run_import_test(&ImportTest {
        name: "NarasimhaProcess/UserTracking",
        set_up_script: &[],
        sql_filename: "NarasimhaProcess_UserTracking.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn nathalyholguin16_sistema_de_gesti_n_de_cine() {
    run_import_test(&ImportTest {
        name: "NathalyHolguin16/Sistema_de_gesti-n_de_Cine",
        set_up_script: &[],
        sql_filename: "NathalyHolguin16_Sistema_de_gesti-n_de_Cine.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn necker55_supermarket_shop() {
    run_import_test(&ImportTest {
        name: "NECKER55/supermarket_shop",
        set_up_script: &[],
        sql_filename: "NECKER55_supermarket_shop.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn nxtrm_neanote() {
    run_import_test(&ImportTest {
        name: "nxtrm/neanote",
        set_up_script: &[],
        sql_filename: "nxtrm_neanote.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn nyfagel_klubb() {
    run_import_test(&ImportTest {
        name: "nyfagel/klubb",
        set_up_script: &[],
        sql_filename: "nyfagel_klubb.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn oknosoft_windowbuilder_planning() {
    run_import_test(&ImportTest {
        name: "oknosoft/windowbuilder-planning",
        set_up_script: &[],
        sql_filename: "oknosoft_windowbuilder-planning.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn openeventdatabase_backend() {
    run_import_test(&ImportTest {
        name: "openeventdatabase/backend",
        set_up_script: &[],
        sql_filename: "openeventdatabase_backend.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn openlawnz_openlawnz_data_processor() {
    run_import_test(&ImportTest {
        name: "openlawnz/openlawnz-data-processor",
        set_up_script: &[],
        sql_filename: "openlawnz_openlawnz-data-processor.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn oslabs_beta_ditto() {
    run_import_test(&ImportTest {
        name: "oslabs-beta/ditto",
        set_up_script: &[],
        sql_filename: "oslabs-beta_ditto.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn paulshriner_fcc_rd_cert() {
    run_import_test(&ImportTest {
        name: "paulshriner/fcc-rd-cert",
        set_up_script: &[],
        sql_filename: "paulshriner_fcc-rd-cert.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn qqtati_diplom() {
    run_import_test(&ImportTest {
        name: "qqtati/diplom",
        set_up_script: &[],
        sql_filename: "qqtati_diplom.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn riclolsen_json_scada() {
    run_import_test(&ImportTest {
        name: "riclolsen/json-scada",
        set_up_script: &[],
        sql_filename: "riclolsen_json-scada.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn rmarquez123_titans() {
    run_import_test(&ImportTest {
        name: "rmarquez123/titans",
        set_up_script: &[],
        sql_filename: "rmarquez123_titans.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn roboflow_scavenger_hunt() {
    run_import_test(&ImportTest {
        name: "roboflow/scavenger-hunt",
        set_up_script: &[
            "CREATE USER supabase_admin WITH SUPERUSER PASSWORD 'password';",
            "CREATE USER anon;",
            "CREATE USER authenticated;",
            "CREATE USER service_role;",
            "CREATE USER supabase_auth_admin;",
            "CREATE USER dashboard_user;",
            "CREATE USER readonly;",
            "CREATE USER partner_token_terminal;",
        ],
        sql_filename: "roboflow_scavenger-hunt.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn rsna_isn_edge_server_database() {
    run_import_test(&ImportTest {
        name: "RSNA/isn-edge-server-database",
        set_up_script: &[],
        sql_filename: "RSNA_isn-edge-server-database.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn s0mbre_russtat() {
    run_import_test(&ImportTest {
        name: "S0mbre/russtat",
        set_up_script: &[],
        sql_filename: "S0mbre_russtat.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn slsfi_digital_edition_db() {
    run_import_test(&ImportTest {
        name: "slsfi/digital_edition_db",
        set_up_script: &[],
        sql_filename: "slsfi_digital_edition_db.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn soniamarrocco_library_database() {
    run_import_test(&ImportTest {
        name: "SoniaMarrocco/Library-Database",
        set_up_script: &[],
        sql_filename: "SoniaMarrocco_Library-Database.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn strangegoofy_yota_game() {
    run_import_test(&ImportTest {
        name: "StrangeGoofy/Yota_game",
        set_up_script: &[],
        sql_filename: "StrangeGoofy_Yota_game.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn stronglogicsolutions_kserver() {
    run_import_test(&ImportTest {
        name: "StronglogicSolutions/kserver",
        set_up_script: &[],
        sql_filename: "StronglogicSolutions_kserver.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn surgefm_v2land_redstone() {
    run_import_test(&ImportTest {
        name: "surgefm/v2land-redstone",
        set_up_script: &[],
        sql_filename: "surgefm_v2land-redstone.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn sylvain_guehria_stockshop() {
    run_import_test(&ImportTest {
        name: "sylvain-guehria/StockShop",
        set_up_script: &[],
        sql_filename: "sylvain-guehria_StockShop.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn tarapadilla_marketspring() {
    run_import_test(&ImportTest {
        name: "TaraPadilla/MarketSpring",
        set_up_script: &[],
        sql_filename: "TaraPadilla_MarketSpring.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn the_benchmarker_web_frameworks() {
    run_import_test(&ImportTest {
        name: "the-benchmarker/web-frameworks",
        set_up_script: &[],
        sql_filename: "the-benchmarker_web-frameworks.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn theophoric_prisma_near_indexer() {
    run_import_test(&ImportTest {
        name: "theophoric/prisma-near-indexer",
        set_up_script: &[
            "CREATE USER testnet WITH SUPERUSER PASSWORD 'password';",
            "CREATE USER cloudsqladmin WITH SUPERUSER PASSWORD 'password';",
            "CREATE USER cloudsqlsuperuser WITH SUPERUSER PASSWORD 'password';",
            "CREATE USER explorer;",
            "CREATE USER wallet;",
            "CREATE USER jupyter;",
            "CREATE USER robertyan;",
            "CREATE USER public_readonly;",
            "CREATE USER readonly;",
            "CREATE USER partner_token_terminal;",
        ],
        sql_filename: "theophoric_prisma-near-indexer.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn timovski_co_opminesweeper() {
    run_import_test(&ImportTest {
        name: "Timovski/Co-opMinesweeper",
        set_up_script: &[],
        sql_filename: "Timovski_Co-opMinesweeper.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn tpolecat_cofree() {
    run_import_test(&ImportTest {
        name: "tpolecat/cofree",
        set_up_script: &[],
        sql_filename: "tpolecat_cofree.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn uzurpastor_uniworks() {
    run_import_test(&ImportTest {
        name: "uzurpastor/UniWorks",
        set_up_script: &[],
        sql_filename: "uzurpastor_UniWorks.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn vanesor_zenith() {
    run_import_test(&ImportTest {
        name: "Vanesor/zenith",
        set_up_script: &[],
        sql_filename: "Vanesor_zenith.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn vinicius02612_sistema_da_associacao() {
    run_import_test(&ImportTest {
        name: "Vinicius02612/sistema_da_associacao",
        set_up_script: &[],
        sql_filename: "Vinicius02612_sistema_da_associacao.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn whoiskatie_e_hotels() {
    run_import_test(&ImportTest {
        name: "WhoIsKatie/e-Hotels",
        set_up_script: &[],
        sql_filename: "WhoIsKatie_e-Hotels.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn wolfufu_hakaton2025spring() {
    run_import_test(&ImportTest {
        name: "wolfufu/Hakaton2025Spring",
        set_up_script: &[],
        sql_filename: "wolfufu_Hakaton2025Spring.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn xarpunk_demexam() {
    run_import_test(&ImportTest {
        name: "Xarpunk/DemExam",
        set_up_script: &[],
        sql_filename: "Xarpunk_DemExam.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn yase_search_yase_engine() {
    run_import_test(&ImportTest {
        name: "yase-search/yase-engine",
        set_up_script: &[],
        sql_filename: "yase-search_yase-engine.sql",
        skip_queries: &[],
    });
}

#[test]
#[ignore = "most dumps fail on the Go server too"]
fn yesk0_kbtu_database_24_25() {
    run_import_test(&ImportTest {
        name: "Yesk0/KBTU_Database_24-25",
        set_up_script: &[],
        sql_filename: "Yesk0_KBTU_Database_24-25.sql",
        skip_queries: &[],
    });
}
