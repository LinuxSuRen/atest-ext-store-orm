[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/707e613bf0d84c3cb1367f98e4b1e463)](https://app.codacy.com/gh/LinuxSuRen/atest-ext-store-orm/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)

# atest-ext-store-orm
ORM database Store Extension for API Testing

This project provides an ORM-based database store extension for API testing, simplifying data storage and retrieval operations. It supports various databases including SQLite, MySQL, PostgreSQL, TDengine, and others, making it versatile for different testing environments.

## Features
- Simplified database operations using ORM.
- Integration with API testing frameworks.
- Support for multiple databases (SQLite, MySQL, PostgreSQL, TDengine, etc.).

## Usage
To use this extension in your API testing project, follow these steps:
1. Install the necessary dependencies.
2. Configure the database connection settings.
3. Integrate the extension into your API tests.

## Quick MySQL Setup with TiUP Playground

You can quickly set up a MySQL-compatible database using [TiUP Playground](https://docs.pingcap.com/tidb/stable/tiup-playground):

1. **Install TiUP** (if not already installed):

    ```sh
    curl --proto '=https' --tlsv1.2 -sSf https://tiup.io/install.sh | sh
    source ~/.profile
    ```

2. **Start a TiDB (MySQL-compatible) cluster:**

    ```sh
    tiup playground
    ```

    This will launch a local TiDB cluster with default settings.

3. **Connect to TiDB using the MySQL client:**

    ```sh
    mysql -h 127.0.0.1 -P 4000 -u root
    ```

    Now you can create databases, tables, and run SQL queries as you would with MySQL.

For more details, see the [TiUP Playground documentation](https://docs.pingcap.com/tidb/stable/tiup-playground).

## Q&A

Run the command `apt-get install build-essential libsqlite3-dev` if you meet the sqlite errors.
