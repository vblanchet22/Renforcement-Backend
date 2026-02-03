import "reflect-metadata";
import { DataSource } from "typeorm";
import { config } from "dotenv";
import { User } from "../entities/User";

config();

export const AppDataSource = new DataSource({
  type: "postgres",
  host: process.env.DB_HOST || "localhost",
  port: parseInt(process.env.DB_PORT || "5432"),
  username: process.env.DB_USERNAME || "coloc_user",
  password: process.env.DB_PASSWORD || "coloc_password",
  database: process.env.DB_NAME || "coloc_db",
  synchronize: false, // Ne pas utiliser en production
  logging: true,
  entities: [User],
  migrations: ["src/migrations/**/*.ts"],
  subscribers: [],
});
