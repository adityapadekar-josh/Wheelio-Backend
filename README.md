# Wheelio - Vehicle Rental Platform

Wheelio is a web-based vehicle rental platform that connects vehicle owners (Hosts) with individuals looking to rent vehicles (Seekers). It provides an easy-to-use system where users can browse, search, and book vehicles, while hosts can list and manage their rentals seamlessly.

## Tech Stack

The backend of Wheelio is built using **Golang**.

## Prerequisites

Before running the server, make sure the following are set up on your local machine:

1. **Go**: You need to have Go installed. You can download it from [here](https://golang.org/dl/).

2. **PostgreSQL**: The database used by Wheelio is PostgreSQL. Ensure that you have PostgreSQL installed and running on your machine. You can download it from [here](https://www.postgresql.org/download/).

3. **Dependencies**: Make sure to install the necessary Go dependencies by running:

   ```bash
   go mod tidy
   ```

4. **Config File**: You will need to create a `config.yaml` file in the root of your project. This file should contain the following configuration:

   ```yaml
   http_server:
     port: "<port>"

   database:
     host: "<host>"
     port: <port>
     user: "<user>"
     password: "<password>"
     name: "<database_name>"
   ```

## Running the Project

To run the project, use the following command in your terminal, replacing `<path_to_config>` with the path to your `config.yaml` file:

```bash
CONFIG_PATH=<path_to_config> go run ./cmd/main.go
```

This will start the server on specified port and connect to the PostgreSQL database as specified in the `config.yaml`.

## Documentation & API Specifications

For detailed documentation on how to use and interact with the Wheelio platform, refer to the following resources:

- [Project Documentation](https://docs.google.com/document/d/1xJxXwMktZdJV5oL5-BbTV6OijxDA9YYbmJcUkTxlrzo/edit?usp=sharing)
- [Database Design](https://dbdesigner.page.link/NAdzRdjJupoQnrWr7)
- [API Documentation](https://docs.google.com/document/d/1-BbK0nP2Gr1HfhBD4yzl2J26sCiVC2G_GoewYPo6H8U/edit?usp=sharing)
