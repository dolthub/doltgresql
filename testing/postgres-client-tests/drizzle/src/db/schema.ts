import {  jsonb,  integer, pgTable, varchar } from 'drizzle-orm/pg-core';

export const usersTable = pgTable("users", {
    id: integer().primaryKey(),
    name: varchar({ length: 255 }).notNull(),
    age: integer().notNull(),
    email: varchar({ length: 255 }).notNull().unique(),
    render: jsonb('render'),
});
