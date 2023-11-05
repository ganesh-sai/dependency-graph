# Getting Started
Welcome to **dependency-graph**, a Go application designed to manage and retrieve service information stored in a MongoDB database. This guide will walk you through setting up and running the application, along with an overview of its core features.


# Starting the Application

To get started with the dependency-graph application, follow these steps:

## Setup MongoDB
You have two options for setting up MongoDB:
###  Direct MongoDB Setup:
Ensure MongoDB is installed and running.

Configure the MongoDB connection by providing the MONGODB_URL environment variable in your application:

```bash
MONGODB_URL=mongodb://<username>:<password>@<hostname>:<port>/servicesdb
```

Replace `<username>, <password>, <hostname>`, and `<port>` with your MongoDB server details.

### Docker Compose Setup:

Use the docker-compose.yml file, as provided in the [repo](./docker-compose.yml).

Run the Application
Once you've set up MongoDB, you can start the dependency-graph application:

Open a terminal and navigate to the project directory.
Use the following command to run the application:

```bash
go run main.go
```
The application will start and be accessible at http://localhost:8080.

## Available APIs
dependency-graph offers a straightforward RESTful API to interact with services. Here are the core functionalities:

**Create Service**: You can create a new service by making a POST request to /services. Provide the service details in the request body in YAML format.

**Get Service**: Retrieve service details by making a GET request to /services/{name}/{version}. You will receive information about the service and the services consuming it.

**Update Service**: Update service details by making a PUT request to /services/{name}/{version}. Provide the updated service details in the request body.

**Delete Service**: Delete a service by making a DELETE request to /services/{name}/{version. If the service is consumed by other services, you can use forceDelete=true to delete it forcibly.

You can use tools like curl or a REST client to make requests and effectively manage your services with dependency-graph.
