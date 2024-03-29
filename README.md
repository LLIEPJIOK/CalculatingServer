## About CalculationServer

Read this in other languages: [Russian](https://github.com/LLIEPJIOK/CalculatingServer/blob/master/README.ru.md).

CalculationServer is a server that calculates expressions within a specified operation calculation time.

The project also utilizes a [PostgresSQL](https://www.postgresql.org) database, [Bootstrap](https://getbootstrap.com), and [HTMX](https://htmx.org). Bootstrap is employed to create a fast and visually appealing interface, while HTMX is used to send requests and handle them in a convenient manner.

## Getting started
To run the server, you need:
- [Go](https://golang.org/dl) version 1.22 or later.
- [PostgreSQL](https://postgresql.org/download). You may need a VPN to install.

To start working with the server, follow these steps:
1. Clone the repository:
   ```bash
   git clone https://github.com/LLIEPJIOK/CalculatingServer.git
   ```
2. Open the `expressions/database.go` file and change the `port` and `password` to match your PostgreSQL port and password.
3. Open console in project folder:
   - Download required packages
      ```bash
      go mod download
      ```
   - Run the program
      ```bash
      go run main.go
      ```
4. Open [`localhost:8080`](http://localhost:8080) in your browser.

## Code structure
1. `main.go` - file for the server where requests are processed.
2. `expressions/expressions.go` - file for handling expressions.
3. `expressions/database.go` - file for interacting with the PostgreSQL database.
4. `static/` - folder for the visual interface. It contains CSS styles, JS script for clearing the input field after sending an expression, and HTML templates with Bootstrap and HTMX.

## How it works
The program consists of a server that handles all requests and agents, which are goroutines responsible for calculating expressions asynchronously. Goroutines also update their status (e.g., waiting, calculating, etc.) every 5 seconds if they are waiting for an expression, as well as before and after performing calculations.

When user send a request (e.g., open the page, submit an expression, etc.), the server handles it, and there are several possible scenarios:

1. **Request to calculate an expression:**
   
   The server add expression to database, then parses it to ensure its validity.
   - If the expression is valid, the server conveys it to the agents through a channel. At some point, one agent picks it up, calculates the expression, and then updates the result.
   - If the expression is invalid, the server sets a parsing error in the expression status.
   
   After that the server update last expressions.

2. **Request to update the operation calculation time:**

   The server updates this data in the database.

3. **Other requests:**

   The server retrieves information from the database and displays it.

![Working scheme](https://github.com/LLIEPJIOK/CalculatingServer/blob/master/images/WorkingScheme.png)

*Note: Currently, there is no automatic updating of data on the page. To see the changes, you must reload the page.*

## Usage example
![Usage example](https://github.com/LLIEPJIOK/CalculatingServer/blob/master/images/ServerUsage.gif)
