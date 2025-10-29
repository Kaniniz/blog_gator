Blog_gator requires Postgres and Go installed to function

The program can be installed by running "go install github.com/Kaniniz/blog_gator@latest" in your cli.

Blog_gator requires a .gatorconfig.json in your home directory to function. 
The file requires two fields, "db_url" and "current_user_name. Current_user_name can be left empty, but there must be a url to your postgres database. 
Example: 
{
    "db_url": "postgres://postgres:gator@localhost:5432/gator?sslmode=disable",
    "current_user_name": ""
}

To use the program type: blog_gator "command"
example blog_gator users

commands:
login "name"                    Login a already registered user
register "name"                 Register a user and login
reset                           Resets the database
users                           Lists all registered users
agg "timeintervall"             Scrapes followed feeds for the active user
addfeed "blog name" "blog url"  Adds and follows a blog feed
feeds                           Lists all registerd feeds and who added them
follow "feed url"               Follows a feed
following                       Lists all followed feeds for the active user
unfollowfeed "feed url"         Unfollows the specified feed for the active user
browse "amount"                 Lists an amount of entries from feeds that have been scraped, defaults to 2 feeds without specified amount
