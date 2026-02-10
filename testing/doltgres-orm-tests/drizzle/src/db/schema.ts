import {  jsonb,  pgTable,  text } from 'drizzle-orm/pg-core';

export const components = pgTable(
    'components',
    {
        id: text('id').notNull(),
        name: text('name'),
        description: text('description'),
        render: jsonb('render'),
    }
);
