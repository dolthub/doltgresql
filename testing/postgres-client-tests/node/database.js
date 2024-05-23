import mysql from "mysql2";

// import * as pg from 'pg'
// const { Pool } = pg

export class Database {
  constructor(config) {
    this.connection = mysql.createConnection(config);
    this.connection.connect();
  }

  query(sql, args) {
    return new Promise((resolve, reject) => {
      this.connection.query(sql, args, (err, rows) => {
        if (err) return reject(err);
        return resolve(rows);
      });
    });
  }

  close() {
    this.connection.end((err) => {
      if (err) {
        console.error(err);
      } else {
        console.log("db connection closed");
      }
    });
  }
}
