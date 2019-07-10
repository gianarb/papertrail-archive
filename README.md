This project offers a set of utils to download and manage logs from papertrail
archives.

## usage
Look at the [documentation](./Documentation/papertrail-archive.md).

```
$ PAPERTRAIL_ARCHIVE_TOKEN=<your_token> papertrail-archive download --from "2019-06-30T00:23:02+02:00" --to "2019-06-30T3:22:02+02:00" --basedir /tmp --no-gunzip
from: 2019-06-29T22:23:02Z
to: 2019-06-30T01:22:02Z
file will be stored to directory (--basedir to change location): /tmp
The archives that will be downloaded are:
        2019-06-29-22
        2019-06-29-23
        2019-06-30-00
do you wan't to proceed with the download? (y)
y
Archive 2019-06-29-22   --- [--------------------------------------------------------------------]   0%
Archive 2019-06-29-23   --- [--------------------------------------------------------------------]   0%
Archive 2019-06-30-00   --- [--------------------------------------------------------------------]   0%

```

## install

```
go get -u github.com/gianarb/papertrail-archive
```
