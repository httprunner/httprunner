.. default-role:: code

Write testcases
===============

It is recommended to write testcases in `YAML` format.

demo
----

And here is testset example of typical scenario: get `token` at the beginning, and each subsequent requests should take the `token` in the headers.

.. code-block:: yaml

    - config:
        name: "create user testsets."
        variables:
            - user_agent: 'iOS/10.3'
            - device_sn: ${gen_random_string(15)}
            - os_platform: 'ios'
            - app_version: '2.8.6'
        request:
            base_url: http://127.0.0.1:5000
            headers:
                Content-Type: application/json
                device_sn: $device_sn

    - test:
        name: get token
        request:
            url: /api/get-token
            method: POST
            headers:
                user_agent: $user_agent
                device_sn: $device_sn
                os_platform: $os_platform
                app_version: $app_version
            json:
                sign: ${get_sign($user_agent, $device_sn, $os_platform, $app_version)}
        extract:
            - token: content.token
        validate:
            - eq: ["status_code", 200]
            - len_eq: ["content.token", 16]

    - test:
        name: create user which does not exist
        request:
            url: /api/users/1000
            method: POST
            headers:
                token: $token
            json:
                name: "user1"
                password: "123456"
        validate:
            - eq: ["status_code", 201]
            - eq: ["content.success", true]

Function invoke is supported in `YAML/JSON` format testcases, such as `gen_random_string` and `get_sign` above. This mechanism relies on the `debugtak.py` hot plugin, with which we can define functions in `debugtak.py` file, and then functions can be auto discovered and invoked in runtime.

For detailed regulations of writing testcases, you can read the :doc:`quickstart` documents.


Comparator
----------

``HttpRunner`` currently supports the following comparators.

+---------------------------+---------------------------+-------------------------+--------------------------+
| comparator                | Description               | A(check), B(expect)     | examples                 |
+===========================+===========================+=========================+==========================+
| ``eq``, ``==``            | value is equal            | A == B                  | 9 eq 9                   |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``lt``                    | less than                 | A < B                   | 7 lt 8                   |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``le``                    | less than or equals       | A <= B                  | 7 le 8, 8 le 8           |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``gt``                    | greater than              | A > B                   | 8 gt 7                   |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``ge``                    | greater than or equals    | A >= B                  | 8 ge 7, 8 ge 8           |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``ne``                    | not equals                | A != B                  | 6 ne 9                   |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``str_eq``                | string equals             | str(A) == str(B)        | 123 str_eq '123'         |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``len_eq``, ``count_eq``  | length or count equals    | len(A) == B             | | 'abc' len_eq 3         |
|                           |                           |                         | | [1,2] len_eq 2         |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``len_gt``, ``count_gt``  | length greater than       | len(A) > B              | | 'abc' len_gt 2         |
|                           |                           |                         | | [1,2,3] len_gt 2       |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``len_ge``, ``count_ge``  | length greater than       | len(A) >= B             | | 'abc' len_ge 3         |
|                           | or equals                 |                         | | [1,2,3] len_gt 3       |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``len_lt``, ``count_lt``  | length less than          | len(A) < B              | | 'abc' len_lt 4         |
|                           |                           |                         | | [1,2,3] len_lt 4       |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``len_le``, ``count_le``  | length less than          | len(A) <= B             | | 'abc' len_le 3         |
|                           | or equals                 |                         | | [1,2,3] len_le 3       |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``contains``              | contains                  | [1, 2] contains 1       | | 'abc' contains 'a'     |
|                           |                           |                         | | [1,2,3] len_lt 4       |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``contained_by``          | contained by              | A in B                  | | 'a' contained_by 'abc' |
|                           |                           |                         | | 1 contained_by [1,2]   |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``type_match``            | A is instance of B        | isinstance(A, B)        | 123 type_match 'int'     |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``regex_match``           | regex matches             | re.match(B, A)          | 'abcdef' regex 'a\w+d'   |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``startswith``            | starts with               | A.startswith(B) is True | 'abc' startswith 'ab'    |
+---------------------------+---------------------------+-------------------------+--------------------------+
| ``endswith``              | ends with                 | A.endswith(B) is True   | 'abc' endswith 'bc'      |
+---------------------------+---------------------------+-------------------------+--------------------------+


Extraction and Validation
-------------------------

Suppose we get the following HTTP response.

.. code-block:: javascript

    // status code: 200

    // response headers
    {
        "Content-Type": "application/json"
    }

    // response body content
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


In `extract` and `validate`, we can do chain operation to extract data field in HTTP response.

For instance, if we want to get `Content-Type` in response headers, then we can specify `headers.content-type`; if we want to get `first_name` in response content, we can specify `content.person.name.first_name`.

There might be slight difference on list, cos we can use index to locate list item. For example, `Guangzhou` in response content can be specified by `content.person.cities.0`.

.. code-block:: javascript

    // get status code
    status_code

    // get headers field
    headers.content-type

    // get content field
    body.success
    content.success
    text.success
    content.person.name.first_name
    content.person.cities.1


.. code-block:: yaml

    extract:
        - content_type: headers.content-type
        - first_name: content.person.name.first_name
    validate:
        - eq: ["status_code", 200]
        - eq: ["headers.content-type", "application/json"]
        - gt: ["headers.content-length", 40]
        - eq: ["content.success", true]
        - len_eq: ["content.token", 16]


.. _QuickStart: http://