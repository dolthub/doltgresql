import 'dotenv/config';
import { Pool } from 'pg';
import { drizzle } from 'drizzle-orm/node-postgres';
import { components } from "./db/schema";

const params = {
    id: 'test',
    render: null,
    name: 'Test',
    description: null,
};

const connectionString = 'postgres://postgres:password@localhost:5432/postgres';

const pool = new Pool({ connectionString });

const db = drizzle(pool);

await db
    .insert(components)
    .values(params)
