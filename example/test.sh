#!/bin/bash

echo ">>>>>>> Cleaning any running Docker containers <<<<<<<"

docker-compose down

echo ">>>>>>> Starting Docker <<<<<<<"

docker-compose up -d

echo ">>>>>>> Building a new version of migrate <<<<<<<<"

go build -o migrate ../go/.

echo ">>>>>>> Waiting for Docker to start (15 seconds) <<<<<<<"

sleep 15

docker ps

echo ">>>>>>> Reset Database State <<<<<<<"

docker exec -t -i example_management_db_1 mysql -ptest -e 'drop database management;'
docker exec -t -i example_management_db_1 mysql -ptest -e 'create database management;'

docker exec -t -i example_target_db_1 mysql -ptest -e 'drop database test;'
docker exec -t -i example_target_db_1 mysql -ptest -e 'create database test;'

echo ">>>>>>> Creating Management Schema <<<<<<<"
./migrate setup --management

echo ">>>>>>> Showing Management Schema <<<<<<<"

docker exec -t -i example_management_db_1 mysql -ptest -e 'show tables;show create table `migration`; show create table `migration_steps`;' management

echo ">>>>>>> Test Target Database Schema <<<<<<<"
docker exec -t -i example_target_db_1 mysql -ptest -e 'show tables;' test

echo ">>>>>>> Resetting Example Schema <<<<<<<"
cat test_schema/cats.yml > working/animals/cats.yml
cat test_schema/ferrari_models.yml > working/cars/manufacturer/ferrari/ferrari_models.yml
cat test_schema/porsche_invoices.yml > working/cars/manufacturer/porsche/porsche_invoices.yml

echo ">>>>>>> Diffing <<<<<<<"
./migrate diff

echo ">>>>>>> Migrating <<<<<<<"
./migrate sandbox --migrate

echo ">>>>>>> Test Target Database Schema <<<<<<<"
docker exec -t -i example_target_db_1 mysql -ptest -e 'show tables;' test

echo ">>>>>>> Modifying Cats Schema <<<<<<<"
cat test_schema/cats_modified.yml > working/animals/cats.yml

echo ">>>>>>> Diffing <<<<<<<"
./migrate diff

echo ">>>>>>> Migrating Modified Schema <<<<<<<"
./migrate sandbox --migrate

echo ">>>>>>> Modifying Ferrari Models Schema <<<<<<<"
cat test_schema/ferrari_models_modified.yml > working/cars/manufacturer/ferrari/ferrari_models.yml

echo ">>>>>>> Diffing <<<<<<<"
./migrate diff

echo ">>>>>>> Migrating Modified Schema <<<<<<<"
./migrate sandbox --migrate

echo ">>>>>>> Modifying Porsche Models Schema <<<<<<<"
cat test_schema/porsche_invoices_modified.yml > working/cars/manufacturer/porsche/porsche_invoices.yml

echo ">>>>>>> Diffing <<<<<<<"
./migrate diff

echo ">>>>>>> Migrating Modified Schema <<<<<<<"
./migrate sandbox --migrate

echo ">>>>>>> Result <<<<<<<"
docker exec -t -i example_target_db_1 mysql -ptest -e 'show tables;' test
docker exec -t -i example_target_db_1 mysql -ptest -e 'show create table `cats`;' test
docker exec -t -i example_target_db_1 mysql -ptest -e 'show create table `dogs`;' test
docker exec -t -i example_target_db_1 mysql -ptest -e 'show create table `porsche_inventory`;' test
docker exec -t -i example_target_db_1 mysql -ptest -e 'show create table `porsche_invoices`;' test
docker exec -t -i example_target_db_1 mysql -ptest -e 'show create table `ferrari_models`;' test
docker exec -t -i example_target_db_1 mysql -ptest -e 'show create table `ferrari_warehouse`;' test

echo ">>>>>>> Stopping Docker Containers <<<<<<<"
docker-compose down
