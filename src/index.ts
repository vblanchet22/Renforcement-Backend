import "reflect-metadata";
import { AppDataSource } from "./config/data-source";

async function main() {
  try {
    // Initialiser la connexion à la base de données
    await AppDataSource.initialize();
    console.log("✓ Connexion à la base de données établie");

    // Vous pouvez ajouter votre logique serveur ici

  } catch (error) {
    console.error("Erreur lors de la connexion à la base de données:", error);
    process.exit(1);
  }
}

main();
