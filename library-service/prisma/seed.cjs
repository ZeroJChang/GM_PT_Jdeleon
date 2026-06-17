const { PrismaPg } = require("@prisma/adapter-pg");
const { PrismaClient } = require("../dist/src/generated/prisma/client");
const bcrypt = require("bcrypt");

const ADMIN_ID = "e11067e4-7927-442b-89ea-533cf5564609";
const USER_ID = "81b8f134-ba1f-42cb-a834-859b9ac23f09";
const BOOK_ID = "117f2aaf-61ab-4b49-b34a-7b331f6947a8";

async function main() {
  const adapter = new PrismaPg({
    connectionString:
      process.env.DATABASE_URL ??
      "postgresql://postgres:postgres@localhost:5433/library_db?schema=public",
  });

  const prisma = new PrismaClient({
    adapter,
  });

  const passwordHash = await bcrypt.hash("123456", 10);

  await prisma.user.upsert({
    where: { email: "admin@test.com" },
    update: {
      name: "Admin User",
      role: "ADMIN",
      passwordHash,
    },
    create: {
      id: ADMIN_ID,
      name: "Admin User",
      email: "admin@test.com",
      role: "ADMIN",
      passwordHash,
    },
  });

  await prisma.user.upsert({
    where: { email: "user@test.com" },
    update: {
      name: "Normal User",
      role: "USER",
      passwordHash,
    },
    create: {
      id: USER_ID,
      name: "Normal User",
      email: "user@test.com",
      role: "USER",
      passwordHash,
    },
  });

  await prisma.book.upsert({
    where: { isbn: "9780132350884" },
    update: {
      title: "Clean Code",
      author: "Robert C. Martin",
      year: 2008,
      genre: "Software Engineering",
      totalCopies: 5,
      availableCopies: 5,
    },
    create: {
      id: BOOK_ID,
      title: "Clean Code",
      author: "Robert C. Martin",
      isbn: "9780132350884",
      year: 2008,
      genre: "Software Engineering",
      totalCopies: 5,
      availableCopies: 5,
    },
  });

  await prisma.$disconnect();

  console.log("Seed completed successfully");
}

main().catch((error) => {
  console.error("Seed failed", error);
  process.exit(1);
});
