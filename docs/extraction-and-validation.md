# Extraction and Validation

Suppose we get the following HTTP response.

```text
# status code: 200

# response headers
{
    "Content-Type": "application/json"
}

# response body content
{
    "success": False,
    "person": {
        "name": {
            "first_name": "Leo",
            "last_name": "Lee",
        },
        "age": 29,
        "cities": ["Guangzhou", "Shenzhen"]
    }
}
```

In `extract` and `validate`, we can do chain operation to extract data field in HTTP response.

For instance, if we want to get `Content-Type` in response headers, then we can specify `headers.content-type`; if we want to get `first_name` in response content, we can specify `content.person.name.first_name`.

There might be slight difference on list, cos we can use index to locate list item. For example, `Guangzhou` in response content can be specified by `content.person.cities.0`.

```text
# get status code
status_code

# get headers field
headers.content-type

# get content field
body.success
content.success
text.success
content.person.name.first_name
content.person.cities.1
```

```yaml
extract:
    - content_type: headers.content-type
    - first_name: content.person.name.first_name
validate:
    - {"check": "status_code", "comparator": "eq", "expected": 200}
    - {"check": "headers.content-type", "expected": "application/json"}
    - {"check": "headers.content-length", "comparator": "gt", "expected": 40}
    - {"check": "content.success", "comparator": "eq", "expected": True}
    - {"check": "content.token", "comparator": "len_eq", "expected": 16}
```
