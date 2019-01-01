go-tinyserver
=============
tiny server

Usage:
------
File serve on current directory
```sh
cd "$serv_root"
tinyserver

# on another terminal
curl localhost:8080/srv/
```

Specify serve file
```sh
tinyserver -file "$file"

# on another terminal
curl localhost:8080/srv/
```

Install:
--------
```sh
go get github.com/yaeshimo/go-tinyserver/tinyserver
```

License:
--------
MIT
