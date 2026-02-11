import 'dotenv/config';
import { Pool } from 'pg';
import { drizzle } from 'drizzle-orm/node-postgres';
import { usersTable } from "./db/schema";
import { eq } from 'drizzle-orm';

const connectionString = process.env.DATABASE_URL!
const pool = new Pool({ connectionString });

const db = drizzle(pool);

async function main() {
    const user: typeof usersTable.$inferInsert = {
        id: 1,
        name: 'John',
        age: 30,
        render: null,
        email: 'john@example.com',
    };
    await db.insert(usersTable).values(user);
    console.log('New user created!')
    const users = await db.select().from(usersTable);
    console.log('Getting all users from the database: ', users)
    /*
    const users: {
      id: number;
      name: string;
      age: number;
      email: string;
      render: jsonb;
    }[]
    */
    await db
        .update(usersTable)
        .set({
            age: 31,
        })
        .where(eq(usersTable.email, user.email));
    console.log('User info updated!')
}

main();
