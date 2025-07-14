# Gator

Gator is a command-line tool for managing feeds and user interactions.

## Prerequisites

Before running Gator, you'll need to have the following installed:

*   **PostgreSQL**: Gator uses a PostgreSQL database to store user and feed data. You can download PostgreSQL from [here](https://www.postgresql.org/download/).
*   **Go**: Gator is written in Go. You can download and install Go from [here](https://golang.org/doc/install).

## Installation

You can install the `gator` CLI tool using `go install`:

```bash
go install github.com/your-username/gator@latest
```

**Note:** Replace `github.com/your-username/gator` with the actual path to this repository.

After installation, the `gator` executable will be placed in your Go bin directory (e.g., `$GOPATH/bin` or `$HOME/go/bin`), which should be in your system's `PATH`. You can verify the installation by running:

```bash
gator --help
```

## Configuration

Before running Gator, you'll need to create a configuration file. By default, Gator looks for a file named `config.yaml` in the directory where you run the command.

Here's an example `config.yaml` file:

```yaml
database:
  host: localhost
  port: 5432
  user: gator_user
  password: your_password
  dbname: gator_db
```

Make sure to replace `gator_user`, `your_password`, and `gator_db` with your actual PostgreSQL credentials and desired database name.

You'll also need to create the `gator_db` database and a user for Gator in your PostgreSQL instance. You can do this by connecting to your PostgreSQL server (e.g., using `psql`) and running:

```sql
CREATE DATABASE gator_db;
CREATE USER gator_user WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE gator_db TO gator_user;
```

## Running the Program and Available Commands

Once you have PostgreSQL installed, your database configured, and Gator installed, you can start using it.

To see a list of available commands, simply run:

```bash
gator
```

Here are some of the commands you can run:

*   **`gator register <username>`**: Registers a new user account with the specified username.
    *   *Example:* `gator register alice`
*   **`gator login <username>`**: Logs in an existing user with the given username. Many commands require you to be logged in.
    *   *Example:* `gator login alice`
*   **`gator reset`**: Resets the database by deleting all users. Use with caution!
*   **`gator users`**: Lists all registered users. The currently logged-in user will be marked.
*   **`gator agg <time_duration>`**: Aggregates and displays content from followed feeds at a specified interval. `time_duration` should be a Go duration string (e.g., `1s`, `1m`, `1h`). This command will run indefinitely.
    *   *Example:* `gator agg 10m`
*   **`gator addfeed <feed_name> <feed_url>`**: (Requires login) Adds a new feed with a given name and URL to your list of available feeds. You will automatically follow this feed.
    *   *Example:* `gator addfeed "My Tech Blog" "https://example.com/tech-blog/rss.xml"`
*   **`gator feeds`**: Lists all available feeds in the database.
*   **`gator follow <feed_url>`**: (Requires login) Starts following a specific feed by its URL.
    *   *Example:* `gator follow "https://example.com/news/feed.xml"`
*   **`gator following`**: (Requires login) Lists all feeds you are currently following.
*   **`gator unfollow <feed_url>`**: (Requires login) Stops following a specific feed by its URL.
    *   *Example:* `gator unfollow "https://example.com/news/feed.xml"`
*   **`gator browse [limit]`**: (Requires login) Browses and displays the latest posts from your followed feeds. Optionally, you can specify a `limit` to control the number of posts displayed (default is 10).
    *   *Example:* `gator browse` (shows 10 posts)
    *   *Example:* `gator browse 50` (shows up to 50 posts)

For more detailed information on any command, you can use the `help` flag:

```bash
gator <command> --help
```
For example:
```bash
gator login --help
```