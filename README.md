# Spider 疯狂采集器

1. 我只是一个疯狂的采集器，不做别的多余的事情，就是给了我请求，那我就去请求。
2. 请求的结果最多储存10000条在内存中，如果没有订阅消费，那么有新的就会去掉最老的那条。
3. 可能会有消息丢失，这个需要应用层自己保证是否重试，我不管，只管请求。

```php
$client = New Spider();

// publish
$client->publish();

// subscribe
$client->subscribe();
```
# License
MIT
