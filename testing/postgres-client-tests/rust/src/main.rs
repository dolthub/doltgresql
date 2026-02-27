use sqlx::postgres::PgPoolOptions;
use sqlx::types::Uuid;
use sqlx::types::chrono::Utc;
use sqlx::types::chrono::DateTime;
use sqlx::types::chrono::NaiveDate;

#[derive(sqlx::FromRow)]
struct Event {
	id: sqlx::types::Uuid,
	created_at: DateTime<Utc>,
	event_date: Option<NaiveDate>,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
	let mut args = std::env::args();
	let program = args.next().unwrap_or_else(|| "app".to_string());
	let user = args.next().ok_or_else(|| {
		format!("Usage: {program} <USER> <PORT>")
	})?;
	let port: u16 = args.next().ok_or_else(|| {
		format!("Usage: {program} <USER> <PORT>")
	})?.parse()?;
	let database_url = format!("postgresql://{user}:password@localhost:{port}/postgres");
	let pool = PgPoolOptions::new()
		.max_connections(5)
		.connect(&database_url)
		.await?;
 
	let exists: bool = sqlx::query_scalar("SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1);")
		.bind("test_table")
		.fetch_one(&pool)
		.await?;
	println!("exists={exists}");

	sqlx::query("DROP TABLE IF EXISTS users, events;")
		.execute(&pool)
		.await?;

	sqlx::query("CREATE TABLE users (id uuid default gen_random_uuid(), name text, email text);")
		.execute(&pool)
		.await?;

	sqlx::query("INSERT INTO users (name, email) VALUES ($1, $2)")
		.bind("Alice")
		.bind("alice@example.com")
		.execute(&pool)
		.await?;

	let some_uuid: Uuid = sqlx::query_scalar("SELECT id FROM users WHERE email = $1 LIMIT 1")
		.bind("alice@example.com")
		.fetch_one(&pool)
		.await?;

	sqlx::query("UPDATE users SET name = $1 WHERE id = $2")
		.bind("Bob")
		.bind(some_uuid)
		.execute(&pool)
		.await?;

	sqlx::query("CREATE TABLE events (id UUID PRIMARY KEY, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), event_date DATE);")
		.execute(&pool)
		.await?;

	let some_id: Uuid = sqlx::query_scalar("INSERT INTO events (id, event_date) VALUES (gen_random_uuid(), '2026-02-12') RETURNING id;")
		.fetch_one(&pool)
		.await?;

	let __event = sqlx::query_as::<_, Event>("SELECT * FROM events WHERE id = $1")
		.bind(some_id)
		.fetch_one(&pool)
		.await?;

	Ok(())
}
