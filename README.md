# golink
### check for links that return 404 on a webpage

### build the binary
_note: requires Docker_
```
make
```

### run the binary
```
golink https://www.tesla.com <other_urls_here>
```

### output

```
Found 98 unique urls:
- https://www.tesla.com/models
- https://www.tesla.com/model3
...

Failed 1 unique urls:
 - https://3.tesla.com/model3/design
2018/11/12 17:37:32 URL Validation took 3.894980197s
```