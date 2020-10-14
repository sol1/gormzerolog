Do you use [Zerolog](https://github.com/rs/zerolog)?  Do you use [GORM](https://gorm.io)?
Would you like your GORM logs to go through Zerolog?  Then this package is for you!

Loosely based on [the package of (nearly) the same
name](https://github.com/Ahmet-Kaplan/gorm-zerolog) by
[Ahmet-Kaplan](https://github.com/Ahmet-Kaplan), but wildly incompatible due to
changes in GORM, and philosophical differences.


# Usage

To use `gormzerolog` in your GORM database instance, you need to do two things.
Firstly, point the GORM config's `Logger` field at an instance of `gormzerolog.Logger`:

```go
db, err := gorm.Open(..., &gorm.Config{Logger: gormzerolog.Logger{}})
```

This will tell GORM to use gormzerolog for all (well, *almost*[^1] all...) of its logging
needs.  However, by default this will log to a "null" zerolog that doesn't do anything.
In order to actually generate logs, a configured zerolog logger needs to be put into the
database's context:

```go
logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

db = db.WithContext(logger.WithContext(context.Background()))
```

Then, whatever the GORM DB instance wants to log, will go through zerolog.


## Example

```go
package main

import (
	"github.com/sol1/gormzerolog"

  "context"
  "os"

  "github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
)

type User struct {
  gorm.Model

  Name  string
}

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormzerolog.Logger{}})
	if err != nil {
		panic(err)
	}
  logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(zerolog.TraceLevel)
  db = db.WithContext(logger.WithContext(context.Background()))

  db.AutoMigrate(&User{})

  db.Save(&User{Name: "Charlie"})
}
```


# Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md).


# Licence

Unless otherwise stated, everything in this repo is covered by the following
copyright notice:

    Copyright (C) 2020  Sol1 Pty Ltd

    This program is free software: you can redistribute it and/or modify it
    under the terms of the GNU General Public License version 3, as
    published by the Free Software Foundation.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.

-----

[^1]: for some unfathomable reason, certain internal errors in GORM still
  go through the built-in logger.  Presumably someone just referred to the
  wrong variable somewhere.  In normal use, you'll never see it, so just
  pretend it doesn't happen.
