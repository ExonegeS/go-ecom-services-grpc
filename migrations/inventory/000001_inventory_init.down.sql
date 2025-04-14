-- Drop tables in reverse order
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;

-- Remove UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";