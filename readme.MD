# How to build

1 - Clone the repo

```
$ git clone git@github.com:nm-morais/CRDrill.git
```

2 - Cd into it

```
$ cd CRDrill
```

3 - Run the command:

```
$ go build .
```

4 - This creates an executable with name `CRDrill`. Make executable with:

```
$ sudo chmod +x CRDrill
```

5 - Finally, to make it available in your path, copy the binary into one of your `bin` directories (we used `/usr/local/bin`) as an example:

```
$ sudo mv CRDrill /usr/local/bin
```


