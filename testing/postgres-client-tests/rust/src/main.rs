use sqlx::postgres::PgPoolOptions;

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
	let exists: bool = sqlx::query_scalar(
		"SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1);"
	)
	.bind("test_table")
	.fetch_one(&pool)
	.await?;
	println!("exists={exists}");
	Ok(())
}
