go-tinyserver
=============
File server

Usage:
------
File serve on current directory
```sh
cd "$serv_root"
gots

# on another terminal
curl localhost:8080/srv/
```

Specify serve file
```sh
gots -file "$file"

# on another terminal
curl localhost:8080/srv/
```

Install:
--------
```sh
go get github.com/yaeshimo/go-tinyserver/gots
```

License:
--------
MIT
