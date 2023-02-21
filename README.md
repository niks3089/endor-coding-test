# endor-coding-test

[![CI](https://github.com/niks3089/endor-coding-test/actions/workflows/master.yml/badge.svg)](https://github.com/niks3089/endor-coding-test/actions/workflows/master.yml)

This project implements the ObjectDB interface. The DB I am using is redis. The key to access the object is of the form `id::name::kind`.
The reason why I am not using just the `id` which is an unique `uuid` is the requirement to get and list objects by name and kind. If I did use just the `id`, then getting objects by name would've been to get all objects and filter by name. This would be the same with listing objects by kind. Hence using the above pattern, I can filter redis based on regex to fetch by id, name and kind. For name and kind, we see multiple keys which we iterate through to extract the objects.

I am also Storing inside the redis only iff the key has the count of 2 `::` delimter. This is to avoid storing the object if id or name has `::`. I am also using `kind` as a filter to decide the kind of object to initialize to unmarshal the object into after reading from redis.

The tests are idempotent and can be triggered manually by going to (this page)[https://github.com/niks3089/endor-coding-test/actions/workflows/master.yml] and triggering the workflow. The workflow spins up a redis server instance and runs the tests aginast it.
