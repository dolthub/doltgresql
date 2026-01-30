// Copyright 2023 Dolthub, Inc.
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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestSmokeTests(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:  "Simple statements",
			Focus: true,
			SetUpScript: []string{
				`CREATE SCHEMA "drizzle";`,
				`CREATE SEQUENCE drizzle."__drizzle_migrations_id_seq" AS int4;`,
				`CREATE TABLE "__drizzle_migrations" (
  "id" integer NOT NULL DEFAULT (nextval('drizzle.__drizzle_migrations_id_seq')),
  "hash" text NOT NULL,
  "created_at" bigint,
  PRIMARY KEY ("id")
);`,
				`INSERT INTO "__drizzle_migrations" ("hash","created_at") VALUES ('d3445cf0eaeb405a6b4b9c8386188aece144d40ba89b9616175ca0f69229cc51',1767821157311);`,
				`CREATE TABLE "projects" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "models" jsonb,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","id")
);`,
				`CREATE TABLE "agent" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "default_sub_agent_id" varchar(256),
  "context_config_id" varchar(256),
  "models" jsonb,
  "status_updates" jsonb,
  "prompt" text,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "agent_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "artifact_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "props" jsonb,
  "render" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "artifact_components_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "context_configs" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "headers_schema" jsonb,
  "context_variables" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "context_configs_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "credential_references" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "type" varchar(256) NOT NULL,
  "credential_store_id" varchar(256) NOT NULL,
  "retrieval_params" jsonb,
  "tool_id" varchar(256),
  "user_id" varchar(256),
  "created_by" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "credential_references_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE UNIQUE INDEX "credential_references_id_unique" ON "credential_references" ("id");`,
				`CREATE UNIQUE INDEX "credential_references_tool_user_unique" ON "credential_references" ("tool_id", "user_id");`,
				`CREATE TABLE "data_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "props" jsonb,
  "render" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "data_components_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "dataset" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "dataset_item" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "dataset_id" text NOT NULL,
  "input" jsonb NOT NULL,
  "expected_output" jsonb,
  "simulation_agent" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_item_dataset_fk" FOREIGN KEY ("tenant_id","project_id","dataset_id") REFERENCES "dataset" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "dataset_run_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "dataset_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_run_config_dataset_fk" FOREIGN KEY ("tenant_id","project_id","dataset_id") REFERENCES "dataset" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "dataset_run_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "dataset_run_config_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "dataset_run_config_id" text NOT NULL,
  "agent_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_run_config_agent_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "dataset_run_config_agent_relations_dataset_run_config_fk" FOREIGN KEY ("tenant_id","project_id","dataset_run_config_id") REFERENCES "dataset_run_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "evaluation_job_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "job_filters" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_job_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "evaluator" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "prompt" text NOT NULL,
  "schema" jsonb NOT NULL,
  "model" jsonb NOT NULL,
  "pass_criteria" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluator_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "evaluation_job_config_evaluator_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_job_config_id" text NOT NULL,
  "evaluator_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_job_cfg_evaluator_rel_evaluator_fk" FOREIGN KEY ("tenant_id","project_id","evaluator_id") REFERENCES "evaluator" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_job_cfg_evaluator_rel_job_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_job_config_id") REFERENCES "evaluation_job_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "evaluation_run_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "is_active" boolean NOT NULL DEFAULT 'true',
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_run_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "evaluation_suite_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "filters" jsonb,
  "sample_rate" double precision,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_suite_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "evaluation_run_config_evaluation_suite_config_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_run_config_id" text NOT NULL,
  "evaluation_suite_config_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_run_cfg_eval_suite_rel_run_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_run_config_id") REFERENCES "evaluation_run_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_run_cfg_eval_suite_rel_suite_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_suite_config_id") REFERENCES "evaluation_suite_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "evaluation_suite_config_evaluator_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_suite_config_id" text NOT NULL,
  "evaluator_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_suite_cfg_evaluator_rel_evaluator_fk" FOREIGN KEY ("tenant_id","project_id","evaluator_id") REFERENCES "evaluator" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_suite_cfg_evaluator_rel_suite_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_suite_config_id") REFERENCES "evaluation_suite_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "external_agents" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "base_url" text NOT NULL,
  "credential_reference_id" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "external_agents_credential_reference_fk" FOREIGN KEY ("credential_reference_id") REFERENCES "credential_references" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
  CONSTRAINT "external_agents_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "functions" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "input_schema" jsonb,
  "execute_code" text NOT NULL,
  "dependencies" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "functions_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "function_tools" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "function_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "function_tools_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "function_tools_function_fk" FOREIGN KEY ("tenant_id","project_id","function_id") REFERENCES "functions" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "sub_agents" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "prompt" text,
  "conversation_history_config" jsonb DEFAULT '{"mode":"full","limit":50,"maxOutputTokens":4000,"includeInternal":false,"messageTypes":["chat","tool-result"]}'::JSONB,
  "models" jsonb,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agents_agents_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "tools" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "config" jsonb NOT NULL,
  "credential_reference_id" varchar(256),
  "credential_scope" varchar(50) NOT NULL DEFAULT 'project',
  "headers" jsonb,
  "image_url" text,
  "capabilities" jsonb,
  "last_error" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "tools_credential_reference_fk" FOREIGN KEY ("credential_reference_id") REFERENCES "credential_references" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
  CONSTRAINT "tools_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "sub_agent_artifact_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "artifact_component_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","sub_agent_id","id"),
  CONSTRAINT "sub_agent_artifact_components_artifact_component_fk" FOREIGN KEY ("tenant_id","project_id","artifact_component_id") REFERENCES "artifact_components" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_artifact_components_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "sub_agent_data_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "data_component_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "sub_agent_data_components_data_component_fk" FOREIGN KEY ("tenant_id","project_id","data_component_id") REFERENCES "data_components" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_data_components_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "sub_agent_external_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "external_agent_id" varchar(256) NOT NULL,
  "headers" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_external_agent_relations_external_agent_fk" FOREIGN KEY ("tenant_id","project_id","external_agent_id") REFERENCES "external_agents" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_external_agent_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "sub_agent_function_tool_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "function_tool_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_function_tool_relations_function_tool_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","function_tool_id") REFERENCES "function_tools" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_function_tool_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "sub_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "source_sub_agent_id" varchar(256) NOT NULL,
  "target_sub_agent_id" varchar(256),
  "relation_type" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "sub_agent_team_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "target_agent_id" varchar(256) NOT NULL,
  "headers" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_team_agent_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_team_agent_relations_target_agent_fk" FOREIGN KEY ("tenant_id","project_id","target_agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`CREATE TABLE "sub_agent_tool_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "tool_id" varchar(256) NOT NULL,
  "selected_tools" jsonb,
  "headers" jsonb,
  "tool_policies" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_tool_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_tool_relations_tool_fk" FOREIGN KEY ("tenant_id","project_id","tool_id") REFERENCES "tools" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`SELECT DOLT_COMMIT('-Am', 'Applied database migrations');`,
				`SELECT DOLT_BRANCH('default_my-weather-project_main');`,
				`INSERT INTO "__drizzle_migrations" ("hash","created_at") VALUES ('634b9140001f10d551fe0d81ca19050f3cc8af8da1ab6c9b6e93d99f33e5fc84',1768766675586);`,
				`CREATE TABLE "triggers" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "enabled" boolean NOT NULL DEFAULT 'true',
  "input_schema" jsonb,
  "output_transform" jsonb,
  "message_template" text NOT NULL,
  "authentication" jsonb,
  "signing_secret" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "triggers_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
				`SELECT DOLT_COMMIT('-Am', 'Applied database migrations');`,
				`SELECT DOLT_CHECKOUT('default_my-weather-project_main');`,
				`INSERT INTO "projects" ("tenant_id","id","name","description","models","stop_when","created_at","updated_at") VALUES ('default','my-weather-project','Weather Project','Project containing sample agent framework using ','{"base": {"model": "openai/gpt-4o-mini"}}',NULL,'2026-01-22 16:19:32.74','2026-01-22 16:19:32.74');`,
				`INSERT INTO "agent" ("tenant_id","id","project_id","name","description","default_sub_agent_id","context_config_id","models","status_updates","prompt","stop_when","created_at","updated_at") VALUES ('default','weather-agent','my-weather-project','Weather agent',NULL,'weather-assistant',NULL,NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.782','2026-01-22 16:19:32.862');`,
				`INSERT INTO "data_components" ("tenant_id","id","project_id","name","description","props","render","created_at","updated_at") VALUES ('default','weather-forecast','my-weather-project','WeatherForecast','A hourly forecast for the weather at a given location','{"type": "object", "required": ["forecast"], "properties": {"forecast": {"type": "array", "items": {"type": "object", "required": ["time", "temperature", "code"], "properties": {"code": {"type": "number", "description": "Weather code at given time"}, "time": {"type": "string", "description": "The time of current item E.g. 12PM, 1PM"}, "temperature": {"type": "number", "description": "The temperature at given time in Farenheit"}}, "additionalProperties": false}, "description": "The hourly forecast for the weather at a given location"}}, "additionalProperties": false}',NULL,'2026-01-22 16:19:32.773665','2026-01-22 16:19:32.773665');`,
				`INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','geocoder-agent','my-weather-project','weather-agent','Geocoder agent','Specialized agent for converting addresses and location names into geographic coordinates. This agent handles all location-related queries and provides accurate latitude/longitude data for weather lookups.','You are a geocoding specialist that converts addresses, place names, and location descriptions
 into precise geographic coordinates. You help users find the exact location they''re asking about
 and provide the coordinates needed for weather forecasting.

 When users provide:
 - Street addresses
 - City names
 - Landmarks
 - Postal codes
 - General location descriptions

 You should use your geocoding tools to find the most accurate coordinates and provide clear
 information about the location found.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.848333','2026-01-22 16:19:32.848333');`,
				`INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','weather-assistant','my-weather-project','weather-agent','Weather assistant','Main weather assistant that coordinates between geocoding and forecasting services to provide comprehensive weather information. This assistant handles user queries and delegates tasks to specialized sub-agents as needed.','You are a helpful weather assistant that provides comprehensive weather information
 for any location worldwide. You coordinate with specialized agents to:

 1. Convert location names/addresses to coordinates (via geocoder)
 2. Retrieve detailed weather forecasts (via weather forecaster)
 3. Present weather information in a clear, user-friendly format

 When users ask about weather:
 - If they provide a location name or address, delegate to the geocoder first
 - Once you have coordinates, delegate to the weather forecaster
 - Present the final weather information in an organized, easy-to-understand format
 - Include relevant details like temperature, conditions, precipitation, wind, etc.
 - Provide helpful context and recommendations when appropriate

 You have access to weather forecast data components that can enhance your responses
 with structured weather information.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.851804','2026-01-22 16:19:32.851804');`,
				`INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','weather-forecaster','my-weather-project','weather-agent','Weather forecaster','Specialized agent for retrieving detailed weather forecasts and current conditions. This agent focuses on providing accurate, up-to-date weather information using geographic coordinates.','You are a weather forecasting specialist that provides detailed weather information
 including current conditions, forecasts, and weather-related insights.

 You work with precise geographic coordinates to deliver:
 - Current weather conditions
 - Short-term and long-term forecasts
 - Temperature, humidity, wind, and precipitation data
 - Weather alerts and advisories
 - Seasonal and climate information

 Always provide clear, actionable weather information that helps users plan their activities.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.844618','2026-01-22 16:19:32.844618');`,
				`INSERT INTO "tools" ("tenant_id","id","project_id","name","description","config","credential_reference_id","credential_scope","headers","image_url","capabilities","last_error","created_at","updated_at") VALUES ('default','fUI2riwrBVJ6MepT8rjx0','my-weather-project','Forecast weather',NULL,'{"mcp": {"server": {"url": "https://weather-mcp-hazel.vercel.app/mcp"}}, "type": "mcp"}',NULL,'project',NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.748','2026-01-22 16:19:32.748');`,
				`INSERT INTO "tools" ("tenant_id","id","project_id","name","description","config","credential_reference_id","credential_scope","headers","image_url","capabilities","last_error","created_at","updated_at") VALUES ('default','fdxgfv9HL7SXlfynPx8hf','my-weather-project','Geocode address',NULL,'{"mcp": {"server": {"url": "https://weather-mcp-hazel.vercel.app/mcp"}}, "type": "mcp"}',NULL,'project',NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.75','2026-01-22 16:19:32.75');`,
				`INSERT INTO "sub_agent_relations" ("tenant_id","id","project_id","agent_id","source_sub_agent_id","target_sub_agent_id","relation_type","created_at","updated_at") VALUES ('default','0y59hwkkyzml4dq4t1sx8','my-weather-project','weather-agent','weather-assistant','weather-forecaster','delegate','2026-01-22 16:19:32.92219','2026-01-22 16:19:32.92219');`,
				`INSERT INTO "sub_agent_relations" ("tenant_id","id","project_id","agent_id","source_sub_agent_id","target_sub_agent_id","relation_type","created_at","updated_at") VALUES ('default','7ye45uc4j5442ihgqwn6d','my-weather-project','weather-agent','weather-assistant','geocoder-agent','delegate','2026-01-22 16:19:32.925527','2026-01-22 16:19:32.925527');`,
				`INSERT INTO "sub_agent_data_components" ("tenant_id","id","project_id","agent_id","sub_agent_id","data_component_id","created_at") VALUES ('default','689yd78rj16p9880bndfo','my-weather-project','weather-agent','weather-assistant','weather-forecast','2026-01-22 16:19:32.907332');`,
				`INSERT INTO "sub_agent_tool_relations" ("tenant_id","id","project_id","agent_id","sub_agent_id","tool_id","selected_tools","headers","tool_policies","created_at","updated_at") VALUES ('default','4kws0lm8bqi1mkzwbvmz4','my-weather-project','weather-agent','weather-forecaster','fUI2riwrBVJ6MepT8rjx0',NULL,NULL,NULL,'2026-01-22 16:19:32.888','2026-01-22 16:19:32.888');`,
				`INSERT INTO "sub_agent_tool_relations" ("tenant_id","id","project_id","agent_id","sub_agent_id","tool_id","selected_tools","headers","tool_policies","created_at","updated_at") VALUES ('default','ttz1a9tnso0sxim79iphr','my-weather-project','weather-agent','geocoder-agent','fdxgfv9HL7SXlfynPx8hf',NULL,NULL,NULL,'2026-01-22 16:19:32.889','2026-01-22 16:19:32.889');`,
				`SELECT DOLT_COMMIT('-Am', '//Update /manage/tenants/default/project-full/my-weather-project via API');`,
				`UPDATE "tools" SET "updated_at"='2026-01-22 16:19:50.912' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';`,
				`UPDATE "tools" SET "updated_at"='2026-01-22 16:19:50.967' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';`,
				`SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');`,
				`INSERT INTO "evaluator" ("tenant_id","id","project_id","name","description","prompt","schema","model","pass_criteria","created_at","updated_at") VALUES ('default','ubqho5lsm6h7bd3ra8loz','my-weather-project','test','test','test','{"type": "object", "required": ["test"], "properties": {"test": {"type": "string", "description": "test"}}, "additionalProperties": false}','{"model": "anthropic/claude-opus-4-5"}',NULL,'2026-01-22 16:20:07.188','2026-01-22 16:20:07.188');`,
				`SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluators via API');`,
				`UPDATE "tools" SET "updated_at"='2026-01-22 16:20:11.438' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';`,
				`UPDATE "tools" SET "updated_at"='2026-01-22 16:20:11.448' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';`,
				`SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');`,
				`UPDATE "tools" SET "updated_at"='2026-01-22 16:20:17.821' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';`,
				`UPDATE "tools" SET "updated_at"='2026-01-22 16:20:18.082' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';`,
				`SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');`,
				`INSERT INTO "evaluation_job_config" ("tenant_id","id","project_id","job_filters","created_at","updated_at") VALUES ('default','tj06kzjt8ltlyixgfzeao','my-weather-project','{"dateRange": {"endDate": "2026-01-23T04:59:59.999Z", "startDate": "2026-01-21T05:00:00.000Z"}}','2026-01-22 16:20:55.774','2026-01-22 16:20:55.774');`,
				`INSERT INTO "evaluation_job_config_evaluator_relations" ("tenant_id","id","project_id","evaluation_job_config_id","evaluator_id","created_at","updated_at") VALUES ('default','5qk0w692h5ij1sxtohdua','my-weather-project','tj06kzjt8ltlyixgfzeao','ubqho5lsm6h7bd3ra8loz','2026-01-22 16:20:55.781','2026-01-22 16:20:55.781');`,
				`SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-job-configs via API');`,
				`INSERT INTO "evaluation_suite_config" ("tenant_id","id","project_id","filters","sample_rate","created_at","updated_at") VALUES ('default','j5gvgluqzwzhjhycrsnpf','my-weather-project','{"agentIds": ["weather-agent"]}',NULL,'2026-01-22 16:21:19.974','2026-01-22 16:21:19.974');`,
				`INSERT INTO "evaluation_suite_config_evaluator_relations" ("tenant_id","id","project_id","evaluation_suite_config_id","evaluator_id","created_at","updated_at") VALUES ('default','tz51dzynx71gits265e9d','my-weather-project','j5gvgluqzwzhjhycrsnpf','ubqho5lsm6h7bd3ra8loz','2026-01-22 16:21:19.982','2026-01-22 16:21:19.982');`,
				`SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-suite-configs via API');`,
				`INSERT INTO "evaluation_run_config" ("tenant_id","id","project_id","name","description","is_active","created_at","updated_at") VALUES ('default','74pgwrprmea2o7e6avbh7','my-weather-project','test','test',true,'2026-01-22 16:21:20.104','2026-01-22 16:21:20.104');`,
				`INSERT INTO "evaluation_run_config_evaluation_suite_config_relations" ("tenant_id","id","project_id","evaluation_run_config_id","evaluation_suite_config_id","created_at","updated_at") VALUES ('default','plb31qfzw9803g6hbjhef','my-weather-project','74pgwrprmea2o7e6avbh7','j5gvgluqzwzhjhycrsnpf','2026-01-22 16:21:20.111','2026-01-22 16:21:20.111');`,
				`SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-run-configs via API');`,
				`UPDATE "tools" SET "updated_at"='2026-01-22 16:21:23.521' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';`,
				`UPDATE "tools" SET "updated_at"='2026-01-22 16:21:23.771' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';`,
				`SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select dolt_merge('main');",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name:  "Panicking Test",
			Focus: true,
			SetUpScript: []string{
				`CREATE TABLE table1 (
            table1_col1 VARCHAR(256),
            table1_col2 VARCHAR(256),
            table1_col3 VARCHAR(256),
            PRIMARY KEY (table1_col1, table1_col3, table1_col2)
        );`,
				`CREATE TABLE table2 (
            table2_col1 VARCHAR(256),
            table2_col2 VARCHAR(256),
            table2_col3 VARCHAR(256),
            table2_col4 TEXT,
            PRIMARY KEY (table2_col1, table2_col3, table2_col2),
            CONSTRAINT table2_fk FOREIGN KEY (table2_col1, table2_col3, table2_col4) REFERENCES table1 (table1_col1, table1_col3, table1_col2) ON DELETE CASCADE ON UPDATE NO ACTION
        );`,
				`SELECT DOLT_COMMIT('-Am', '1');`,
				`SELECT DOLT_BRANCH('other_branch');`,
				`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
				`SELECT DOLT_COMMIT('-Am', '2');`,
				`SELECT DOLT_CHECKOUT('other_branch');`,
				`INSERT INTO table1 (table1_col1, table1_col2, table1_col3) VALUES ('abc','def','ghi');`,
				`SELECT DOLT_COMMIT('-Am', '3');`,
				`INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','jkl','ghi','def');`,
				`SELECT DOLT_COMMIT('-Am', '4');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select dolt_merge('main');",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Simple statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test VALUES (1, 1), (2, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test2 VALUES (3, 3), (4, 4);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{3, 3},
						{4, 4},
					},
				},
				{
					Query: "SELECT test2.pk FROM test2;",
					Expected: []sql.Row{
						{3},
						{4},
					},
				},
				{
					Query: "SELECT * FROM test ORDER BY 1 LIMIT 1 OFFSET 1;",
					Expected: []sql.Row{
						{2, 2},
					},
				},
				{
					Query:    "SELECT NULL = NULL",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    ";",
					Expected: []sql.Row{},
				},
				{
					Query:    " ; ",
					Expected: []sql.Row{},
				},
				{
					Query:    "-- this is only a comment",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Insert statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 2, 3);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test (v1, pk) VALUES (5, 4);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test (pk, v2) SELECT pk + 5, v2 + 10 FROM test WHERE v2 IS NOT NULL;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, nil},
						{6, nil, 13},
					},
				},
			},
		},
		{
			Name: "Update statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
				"INSERT INTO test VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v2 = 10;",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE test SET v1 = pk + v2;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 11, 10},
						{4, 14, 10},
						{7, 17, 10},
					},
				},
				{
					Query:    "UPDATE test SET pk = subquery.val FROM (SELECT 22 as val) AS subquery WHERE pk >= 7;",
					Skip:     true, // FROM not yet supported
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Skip:  true, // Above query doesn't run yet
					Expected: []sql.Row{
						{1, 11, 10},
						{4, 14, 10},
						{22, 17, 10},
					},
				},
			},
		},
		{
			Name: "Delete statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
				"INSERT INTO test VALUES (1, 1, 1), (2, 3, 4), (5, 7, 9);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DELETE FROM test WHERE v2 = 9;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE v1 = pk;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{2, 3, 4},
					},
				},
			},
		},
		{
			Name: "USE statements",
			SetUpScript: []string{
				"CREATE DATABASE test",
				"USE test",
				"CREATE TABLE t1 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO t1 VALUES (1, 1), (2, 2);",
				"select dolt_commit('-Am', 'initial commit');",
				"select dolt_branch('b1');",
				"select dolt_checkout('b1');",
				"INSERT INTO t1 VALUES (3, 3), (4, 4);",
				"select dolt_commit('-Am', 'commit b1');",
				"select dolt_tag('tag1')",
				"INSERT INTO t1 VALUES (5, 5), (6, 6);",
				"select dolt_checkout('main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from t1 order by 1;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query:            "USE test/b1",
					SkipResultsCheck: true,
				},
				{
					Query: "select * from t1 order by 1;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
						{3, 3},
						{4, 4},
						{5, 5},
						{6, 6},
					},
				},
				{
					Query:            "USE \"test/main\"",
					SkipResultsCheck: true,
				},
				{
					Query: "select * from t1 order by 1;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query:            "USE 'test/tag1'",
					SkipResultsCheck: true,
				},
				{
					Query: "select * from t1 order by 1;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
						{3, 3},
						{4, 4},
					},
				},
			},
		},
		{
			Name: "Boolean results",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT 1 IN (2);",
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query: "SELECT 2 IN (2);",
					Expected: []sql.Row{
						{"t"},
					},
				},
			},
		},
		{
			Name: "Commit and diff across branches",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO test VALUES (1, 1), (2, 2);",
				"SELECT DOLT_ADD('-A');",
				"SELECT DOLT_COMMIT('-m', 'initial commit');",
				"SELECT DOLT_BRANCH('other');",
				"UPDATE test SET v1 = 3;",
				"SELECT DOLT_ADD('-A');",
				"SELECT DOLT_COMMIT('-m', 'commit main');",
				"SELECT DOLT_CHECKOUT('other');",
				"UPDATE test SET v1 = 4 WHERE pk = 2;",
				"SELECT DOLT_ADD('-A');",
				"SELECT DOLT_COMMIT('-m', 'commit other');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT DOLT_CHECKOUT('main');",
					Expected: []sql.Row{{"{0,\"Switched to branch 'main'\"}"}},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 3},
						{2, 3},
					},
				},
				{
					Query:    "SELECT DOLT_CHECKOUT('other');",
					Expected: []sql.Row{{"{0,\"Switched to branch 'other'\"}"}},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 4},
					},
				},
				{
					Query: "SELECT from_pk, to_pk, from_v1, to_v1 FROM dolt_diff_test;",
					Expected: []sql.Row{
						{2, 2, 2, 4},
						{nil, 1, nil, 1},
						{nil, 2, nil, 2},
					},
				},
			},
		},
		{
			Name: "ARRAY expression",
			SetUpScript: []string{
				"CREATE TABLE test1 (id INTEGER primary key, v1 BOOLEAN);",
				"INSERT INTO test1 VALUES (1, 'true'), (2, 'false');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT ARRAY[v1]::boolean[] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t}"},
						{"{f}"},
					},
				},
				{
					Query: "SELECT ARRAY[v1] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t}"},
						{"{f}"},
					},
				},
				{
					Query: "SELECT ARRAY[v1, true, v1] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t,t,t}"},
						{"{f,t,f}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::float8, 2::numeric];",
					Expected: []sql.Row{
						{"{1,2}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::float8, NULL];",
					Expected: []sql.Row{
						{"{1,NULL}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::int2, 2::int4, 3::int8]::varchar[];",
					Expected: []sql.Row{
						{"{1,2,3}"},
					},
				},
				{
					Query:       "SELECT ARRAY[1::int8]::int;",
					ExpectedErr: "cast from `bigint[]` to `integer` does not exist",
				},
				{
					Query:       "SELECT ARRAY[1::int8, 2::varchar];",
					ExpectedErr: "ARRAY types bigint and varchar cannot be matched",
				},
			},
		},
		{
			Name: "Array casting",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT '{true,false,true}'::boolean[];`,
					Expected: []sql.Row{
						{`{t,f,t}`},
					},
				},
				{
					Skip:  true, // TODO: result differs from Postgres
					Query: `SELECT '{"\x68656c6c6f", "\x776f726c64", "\x6578616d706c65"}'::bytea[]::text[];`,
					Expected: []sql.Row{
						{`{"\\x7836383635366336633666","\\x7837373666373236633634","\\x783635373836313664373036633635"}`},
					},
				},
				{
					Skip:  true, // TODO: result differs from Postgres
					Query: `SELECT '{"\\x68656c6c6f", "\\x776f726c64", "\\x6578616d706c65"}'::bytea[]::text[];`,
					Expected: []sql.Row{
						{`{"\\x68656c6c6f", "\\x776f726c64", "\\x6578616d706c65"}`},
					},
				},
				{
					Query: `SELECT '{"abcd", "efgh", "ijkl"}'::char(3)[];`,
					Expected: []sql.Row{
						{`{abc,efg,ijk}`},
					},
				},
				{
					Query: `SELECT '{"2020-02-03", "2020-04-05", "2020-06-06"}'::date[];`,
					Expected: []sql.Row{
						{`{2020-02-03,2020-04-05,2020-06-06}`},
					},
				},
				{
					Query: `SELECT '{1.25,2.5,3.75}'::float4[];`,
					Expected: []sql.Row{
						{`{1.25,2.5,3.75}`},
					},
				},
				{
					Query: `SELECT '{4.25,5.5,6.75}'::float8[];`,
					Expected: []sql.Row{
						{`{4.25,5.5,6.75}`},
					},
				},
				{
					Query: `SELECT '{1,2,3}'::int2[];`,
					Expected: []sql.Row{
						{`{1,2,3}`},
					},
				},
				{
					Query: `SELECT '{4,5,6}'::int4[];`,
					Expected: []sql.Row{
						{`{4,5,6}`},
					},
				},
				{
					Query: `SELECT '{7,8,9}'::int8[];`,
					Expected: []sql.Row{
						{`{7,8,9}`},
					},
				},
				{
					Query: `SELECT '{"{\"a\":\"val1\"}", "{\"b\":\"value2\"}", "{\"c\": \"object_value3\"}"}'::json[];`,
					Expected: []sql.Row{
						{`{"{\"a\":\"val1\"}","{\"b\":\"value2\"}","{\"c\": \"object_value3\"}"}`},
					},
				},
				{
					Query: `SELECT '{"{\"d\":\"val1\"}", "{\"e\":\"value2\"}", "{\"f\": \"object_value3\"}"}'::jsonb[];`,
					Expected: []sql.Row{
						{`{"{\"d\": \"val1\"}","{\"e\": \"value2\"}","{\"f\": \"object_value3\"}"}`},
					},
				},
				{
					Query: `SELECT '{"the", "legendary", "formula"}'::name[];`,
					Expected: []sql.Row{
						{`{the,legendary,formula}`},
					},
				},
				{
					Query: `SELECT '{10.01,20.02,30.03}'::numeric[];`,
					Expected: []sql.Row{
						{`{10.01,20.02,30.03}`},
					},
				},
				{
					Query: `SELECT '{1,10,100}'::oid[];`,
					Expected: []sql.Row{
						{`{1,10,100}`},
					},
				},
				{
					Query: `SELECT '{"this", "is", "some", "text"}'::text[], '{text,without,quotes}'::text[], '{null,NULL,"NULL","quoted"}'::text[];`,
					Expected: []sql.Row{
						{`{this,is,some,text}`, `{text,without,quotes}`, `{NULL,NULL,"NULL",quoted}`},
					},
				},
				{
					Query: `SELECT '{"12:12:13", "14:14:15", "16:16:17"}'::time[];`,
					Expected: []sql.Row{
						{`{12:12:13,14:14:15,16:16:17}`},
					},
				},
				{
					Query: `SELECT '{"2020-02-03 12:13:14", "2020-04-05 15:16:17", "2020-06-06 18:19:20"}'::timestamp[];`,
					Expected: []sql.Row{
						{`{"2020-02-03 12:13:14","2020-04-05 15:16:17","2020-06-06 18:19:20"}`},
					},
				},
				{
					Query: `SELECT '{"3920fd79-7b53-437c-b647-d450b58b4532", "a594c217-4c63-4669-96ec-40eed180b7cf", "4367b70d-8d8b-4969-a1aa-bf59536455fb"}'::uuid[];`,
					Expected: []sql.Row{
						{`{3920fd79-7b53-437c-b647-d450b58b4532,a594c217-4c63-4669-96ec-40eed180b7cf,4367b70d-8d8b-4969-a1aa-bf59536455fb}`},
					},
				},
				{
					Query: `SELECT '{"somewhere", "over", "the", "rainbow"}'::varchar(5)[];`,
					Expected: []sql.Row{
						{`{somew,over,the,rainb}`},
					},
				},
				{
					Query: `SELECT '{1,2,3}'::xid[];`,
					Expected: []sql.Row{
						{`{1,2,3}`},
					},
				},
				{
					Query:       `SELECT '{"abc""","def"}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,c'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT 'a,b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{"a,b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a",b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,"c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,c"}'::text[];`,
					ExpectedErr: "malformed",
				},
			},
		},
		{
			Name: "BETWEEN",
			SetUpScript: []string{
				"CREATE TABLE test (v1 FLOAT8);",
				"INSERT INTO test VALUES (1), (3), (7);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query:    "SELECT * FROM test WHERE v1 BETWEEN 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
			},
		},
		{
			Name: "IN",
			SetUpScript: []string{
				"CREATE TABLE test(v1 INT4, v2 INT4);",
				"INSERT INTO test VALUES (1, 1), (2, 2), (3, 3), (4, 4), (5, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{2, 2},
						{3, 3},
						{4, 4},
					},
				},
				{
					Query:    "CREATE INDEX v2_idx ON test(v2);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v2 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{2, 2},
						{3, 3},
						{4, 4},
					},
				},
			},
		},
		{
			Name: "SUM",
			SetUpScript: []string{
				"CREATE TABLE test(pk SERIAL PRIMARY KEY, v1 INT4);",
				"INSERT INTO test (v1) VALUES (1), (2), (3), (4), (5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT SUM(v1) FROM test WHERE v1 BETWEEN 3 AND 5;",
					Expected: []sql.Row{
						{12.0},
					},
				},
				{
					Query:    "CREATE INDEX v1_idx ON test(v1);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT SUM(v1) FROM test WHERE v1 BETWEEN 3 AND 5;",
					Expected: []sql.Row{
						{12.0},
					},
				},
			},
		},
		{
			Name: "Empty statement",
			Assertions: []ScriptTestAssertion{
				{
					Query:    ";",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Unsupported MySQL statements",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SHOW CREATE TABLE;",
					ExpectedErr: "syntax error",
				},
			},
		},
		{
			Name: "querying tables with same name as pg_catalog tables",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT attname FROM pg_catalog.pg_attribute ORDER BY attname LIMIT 3;",
					Expected: []sql.Row{
						{"ACTION_CONDITION"},
						{"ACTION_ORDER"},
						{"ACTION_ORIENTATION"},
					},
				},
				{
					Query: "SELECT attname FROM pg_attribute ORDER BY attname LIMIT 3;",
					Expected: []sql.Row{
						{"ACTION_CONDITION"},
						{"ACTION_ORDER"},
						{"ACTION_ORIENTATION"},
					},
				},
				{
					Query:    "CREATE TABLE pg_attribute (id INT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "insert into pg_attribute values (1);",
					ExpectedErr: "number of values does not match number of columns provided",
				},
				{
					Query:    "insert into public.pg_attribute values (1);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT attname FROM pg_attribute ORDER BY attname LIMIT 3;",
					Expected: []sql.Row{
						{"ACTION_CONDITION"},
						{"ACTION_ORDER"},
						{"ACTION_ORIENTATION"},
					},
				},
				{
					Query:    "SELECT * FROM public.pg_attribute;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:       "drop table pg_attribute;",
					ExpectedErr: "tables cannot be dropped on database pg_catalog",
				},
				{
					Query:    "drop table public.pg_attribute;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT * FROM public.pg_attribute;",
					ExpectedErr: "table not found: pg_attribute",
				},
			},
		},
		{
			Name: "200 Row Test",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY);",
				"INSERT INTO test VALUES " +
					"(1),   (2),   (3),   (4),   (5),   (6),   (7),   (8),   (9),   (10)," +
					"(11),  (12),  (13),  (14),  (15),  (16),  (17),  (18),  (19),  (20)," +
					"(21),  (22),  (23),  (24),  (25),  (26),  (27),  (28),  (29),  (30)," +
					"(31),  (32),  (33),  (34),  (35),  (36),  (37),  (38),  (39),  (40)," +
					"(41),  (42),  (43),  (44),  (45),  (46),  (47),  (48),  (49),  (50)," +
					"(51),  (52),  (53),  (54),  (55),  (56),  (57),  (58),  (59),  (60)," +
					"(61),  (62),  (63),  (64),  (65),  (66),  (67),  (68),  (69),  (70)," +
					"(71),  (72),  (73),  (74),  (75),  (76),  (77),  (78),  (79),  (80)," +
					"(81),  (82),  (83),  (84),  (85),  (86),  (87),  (88),  (89),  (90)," +
					"(91),  (92),  (93),  (94),  (95),  (96),  (97),  (98),  (99),  (100)," +
					"(101), (102), (103), (104), (105), (106), (107), (108), (109), (110)," +
					"(111), (112), (113), (114), (115), (116), (117), (118), (119), (120)," +
					"(121), (122), (123), (124), (125), (126), (127), (128), (129), (130)," +
					"(131), (132), (133), (134), (135), (136), (137), (138), (139), (140)," +
					"(141), (142), (143), (144), (145), (146), (147), (148), (149), (150)," +
					"(151), (152), (153), (154), (155), (156), (157), (158), (159), (160)," +
					"(161), (162), (163), (164), (165), (166), (167), (168), (169), (170)," +
					"(171), (172), (173), (174), (175), (176), (177), (178), (179), (180)," +
					"(181), (182), (183), (184), (185), (186), (187), (188), (189), (190)," +
					"(191), (192), (193), (194), (195), (196), (197), (198), (199), (200);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY pk;",
					Expected: []sql.Row{
						{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10},
						{11}, {12}, {13}, {14}, {15}, {16}, {17}, {18}, {19}, {20},
						{21}, {22}, {23}, {24}, {25}, {26}, {27}, {28}, {29}, {30},
						{31}, {32}, {33}, {34}, {35}, {36}, {37}, {38}, {39}, {40},
						{41}, {42}, {43}, {44}, {45}, {46}, {47}, {48}, {49}, {50},
						{51}, {52}, {53}, {54}, {55}, {56}, {57}, {58}, {59}, {60},
						{61}, {62}, {63}, {64}, {65}, {66}, {67}, {68}, {69}, {70},
						{71}, {72}, {73}, {74}, {75}, {76}, {77}, {78}, {79}, {80},
						{81}, {82}, {83}, {84}, {85}, {86}, {87}, {88}, {89}, {90},
						{91}, {92}, {93}, {94}, {95}, {96}, {97}, {98}, {99}, {100},
						{101}, {102}, {103}, {104}, {105}, {106}, {107}, {108}, {109}, {110},
						{111}, {112}, {113}, {114}, {115}, {116}, {117}, {118}, {119}, {120},
						{121}, {122}, {123}, {124}, {125}, {126}, {127}, {128}, {129}, {130},
						{131}, {132}, {133}, {134}, {135}, {136}, {137}, {138}, {139}, {140},
						{141}, {142}, {143}, {144}, {145}, {146}, {147}, {148}, {149}, {150},
						{151}, {152}, {153}, {154}, {155}, {156}, {157}, {158}, {159}, {160},
						{161}, {162}, {163}, {164}, {165}, {166}, {167}, {168}, {169}, {170},
						{171}, {172}, {173}, {174}, {175}, {176}, {177}, {178}, {179}, {180},
						{181}, {182}, {183}, {184}, {185}, {186}, {187}, {188}, {189}, {190},
						{191}, {192}, {193}, {194}, {195}, {196}, {197}, {198}, {199}, {200},
					},
				},
			},
		},
		{
			Name: "INDEX as column name",
			SetUpScript: []string{
				`CREATE TABLE test1 (index INT4, CONSTRAINT index_constraint1 CHECK ((index >= 0)));`,
				`CREATE TABLE test2 ("IndeX" INT4, CONSTRAINT index_constraint2 CHECK (("IndeX" >= 0)));`,
				`INSERT INTO test1 VALUES (1);`,
				`INSERT INTO test2 VALUES (2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT * FROM test1;`,
					ExpectedColNames: []string{"index"},
					Expected:         []sql.Row{{1}},
				},
				{
					Query:            `SELECT * FROM test2;`,
					ExpectedColNames: []string{"IndeX"},
					Expected:         []sql.Row{{2}},
				},
				{
					Query:       `INSERT INTO test1 VALUES (-1);`,
					ExpectedErr: "index_constraint1",
				},
				{
					Query:       `INSERT INTO test2 VALUES (-1);`,
					ExpectedErr: "index_constraint2",
				},
			},
		},
	})
}

func TestEmptyQuery(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			// TODO: we want to be able to assert that the empty query returns a specific postgres backend message,
			//  EmptyQueryResponse. The pg library automatically converts this response to an empty-string CommandTag,
			//  which we can't tell apart from other empty CommandTag responses. We do assert that the command tag is empty,
			//  but it would nice to be able to assert a particular message type.
			Name: "Empty query test",
			Assertions: []ScriptTestAssertion{
				{
					Query:       ";",
					ExpectedTag: EmptyCommandTag,
				},
				{
					Query:       " ",
					ExpectedTag: EmptyCommandTag,
				},
			},
		},
	})
}
