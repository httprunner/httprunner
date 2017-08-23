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

In `extract_binds` and `validators`, we can do chain operation to extract data field in HTTP response.

For instance, if we want to get `Content-Type` in response headers, then we can specify `headers.content-type`; if we want to get `first_name` in response content, we can specify `content.person.name.first_name`.

There might be slight difference on list, cos we can use index to locate list item. For example, `Guangzhou` in response content can be specified by `content.person.cities.0`.

```text
{"resp_status_code": "status_code"},
{"resp_headers_content_type": "headers.content-type"},
{"resp_content_body_success": "body.success"},
{"resp_content_content_success": "content.success"},
{"resp_content_text_success": "text.success"},
{"resp_content_person_first_name": "content.person.name.first_name"},
{"resp_content_cities_1": "content.person.cities.1"}
```

```yaml
validators:
    - {"check": "status_code", "comparator": "eq", "expected": 200}
    - {"check": "headers.content-type", "expected": "application/json"}
    - {"check": "headers.content-length", "comparator": "gt", "expected": 40}
    - {"check": "content.success", "comparator": "eq", "expected": True}
    - {"check": "content.token", "comparator": "len_eq", "expected": 16}
```
