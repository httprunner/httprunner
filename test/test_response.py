import requests
from ate import response, exception
from test.base import ApiServerUnittest

class TestResponse(ApiServerUnittest):

    def test_parse_response_object_json(self):
        url = "http://127.0.0.1:5000/api/users"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)
        parsed_dict = resp_obj.parsed_dict()
        self.assertIn('status_code', parsed_dict)
        self.assertIn('headers', parsed_dict)
        self.assertIn('body', parsed_dict)
        self.assertIn('Content-Type', parsed_dict['headers'])
        self.assertIn('Content-Length', parsed_dict['headers'])
        self.assertIn('success', parsed_dict['body'])

    def test_parse_response_object_text(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)
        parsed_dict = resp_obj.parsed_dict()
        self.assertIn('status_code', parsed_dict)
        self.assertIn('headers', parsed_dict)
        self.assertIn('body', parsed_dict)
        self.assertIn('Content-Type', parsed_dict['headers'])
        self.assertIn('Content-Length', parsed_dict['headers'])
        self.assertTrue(str, type(parsed_dict['body']))

    def test_extract_response_json(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'headers': {
                    'Content-Type': "application/json"
                },
                'body': {
                    'success': False,
                    "person": {
                        "name": {
                            "first_name": "Leo",
                            "last_name": "Lee",
                        },
                        "age": 29,
                        "cities": ["Guangzhou", "Shenzhen"]
                    }
                }
            }
        )

        extract_binds = {
            "resp_status_code": "status_code",
            "resp_headers_content_type": "headers.content-type",
            "resp_content_body_success": "body.success",
            "resp_content_content_success": "content.success",
            "resp_content_text_success": "text.success",
            "resp_content_person_first_name": "content.person.name.first_name",
            "resp_content_cities_1": "content.person.cities.1"
        }
        resp_obj = response.ResponseObject(resp)
        extract_binds_dict = resp_obj.extract_response(extract_binds)

        self.assertEqual(
            extract_binds_dict["resp_status_code"],
            200
        )
        self.assertEqual(
            extract_binds_dict["resp_headers_content_type"],
            "application/json"
        )
        self.assertEqual(
            extract_binds_dict["resp_content_content_success"],
            False
        )
        self.assertEqual(
            extract_binds_dict["resp_content_text_success"],
            False
        )
        self.assertEqual(
            extract_binds_dict["resp_content_person_first_name"],
            "Leo"
        )
        self.assertEqual(
            extract_binds_dict["resp_content_cities_1"],
            "Shenzhen"
        )

    def test_extract_response_fail(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'headers': {
                    'Content-Type': "application/json"
                },
                'body': {
                    'success': False,
                    "person": {
                        "name": {
                            "first_name": "Leo",
                            "last_name": "Lee",
                        },
                        "age": 29,
                        "cities": ["Guangzhou", "Shenzhen"]
                    }
                }
            }
        )

        extract_binds = {
            "resp_content_dict_key_error": "content.not_exist"
        }
        resp_obj = response.ResponseObject(resp)

        with self.assertRaises(exception.ParamsError):
            resp_obj.extract_response(extract_binds)

        extract_binds = {
            "resp_content_list_index_error": "content.person.cities.3"
        }
        resp_obj = response.ResponseObject(resp)

        with self.assertRaises(exception.ParamsError):
            resp_obj.extract_response(extract_binds)

    def test_extract_response_json_string(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'headers': {
                    'Content-Type': "application/json"
                },
                'body': "abc"
            }
        )

        extract_binds = {
            "resp_content_body": "content"
        }
        resp_obj = response.ResponseObject(resp)

        extract_binds_dict = resp_obj.extract_response(extract_binds)
        self.assertEqual(
            extract_binds_dict["resp_content_body"],
            "abc"
        )

    def test_validate(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)

        validators = [
            {"check": "resp_status_code", "comparator": "eq", "expected": 201},
            {"check": "resp_body_success", "comparator": "eq", "expected": True}
        ]
        variables_mapping = {
            "resp_status_code": 200,
            "resp_body_success": True
        }

        diff_content_list = resp_obj.validate(validators, variables_mapping)
        self.assertFalse(resp_obj.success)
        self.assertEqual(
            diff_content_list,
            [
                {
                    "check": "resp_status_code",
                    "comparator": "eq", "expected": 201, "value": 200
                }
            ]
        )

        validators = [
            {"check": "resp_status_code", "comparator": "eq", "expected": 201},
            {"check": "resp_body_success", "comparator": "eq", "expected": True}
        ]
        variables_mapping = {
            "resp_status_code": 201,
            "resp_body_success": True
        }

        diff_content_list = resp_obj.validate(validators, variables_mapping)
        self.assertTrue(resp_obj.success)
        self.assertEqual(diff_content_list, [])

    def test_validate_exception(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)

        # expected value missed in validators
        validators = [
            {"check": "status_code", "comparator": "eq", "expected": 201},
            {"check": "body_success", "comparator": "eq"}
        ]
        variables_mapping = {}
        with self.assertRaises(exception.ParamsError):
            resp_obj.validate(validators, variables_mapping)

        # expected value missed in variables mapping
        validators = [
            {"check": "resp_status_code", "comparator": "eq", "expected": 201},
            {"check": "body_success", "comparator": "eq"}
        ]
        variables_mapping = {
            "resp_status_code": 200
        }
        with self.assertRaises(exception.ParamsError):
            resp_obj.validate(validators, variables_mapping)
