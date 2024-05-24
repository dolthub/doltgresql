import pkg from 'pg';
const { Client } = pkg;

export class Database {
  constructor(config) {
    this.client = new Client(config);
    this.client.connect();
  }

  query(sql, args) {
    return new Promise((resolve, reject) => {
      this.client.query(sql, args, (err, rows) => {
        if (err) return reject(err);
        return resolve(rows);
      });
    });
  }

  close() {
    this.client.end((err) => {
      if (err) {
        console.error(err);
      } else {
        console.log("db connection closed");
      }
    });
  }
}
