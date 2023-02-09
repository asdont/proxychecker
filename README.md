# API Proxy Checker

## Directories

---

**./configs/conf.toml** - configuration

**./data/geo-db.bin** - geo DB

**./docs** - OpenAPI

## Build/Run

---

**build/run**

```sh
 make build
 ```

``` sh
./apiproxychecker
```

```
http://localhost:30122
```

```
http://localhost:30122/doc/index.html
```

**run only**

``` sh
make run
```

```
http://localhost:30122
```

```
http://localhost:30122/doc/index.html
```

## Docker-compose

---

```sh
make d.c.build
```

```sh
make d.c.run
```

```
http://localhost:30122
```

```
http://localhost:30122/doc/index.html
```

## Interface

---

>![app settings](docs/screen.jpg)

---