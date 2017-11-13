import requests
from httprunner import response, exception
from tests.base import ApiServerUnittest

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

        extract_binds_list = [
            {"resp_status_code": "status_code"},
            {"resp_headers_content_type": "headers.content-type"},
            {"resp_content_body_success": "body.success"},
            {"resp_content_content_success": "content.success"},
            {"resp_content_text_success": "text.success"},
            {"resp_content_person_first_name": "content.person.name.first_name"},
            {"resp_content_cities_1": "content.person.cities.1"}
        ]
        resp_obj = response.ResponseObject(resp)
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)

        self.assertEqual(
            extract_binds_dict["resp_status_code"],
            200
        )
        self.assertEqual(
            extract_binds_dict["resp_headers_content_type"],
            "application/json"
        )
        self.assertEqual(
            extract_binds_dict["resp_content_body_success"],
            False
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

        extract_binds_list = [
            {"resp_content_dict_key_error": "content.not_exist"}
        ]
        resp_obj = response.ResponseObject(resp)

        with self.assertRaises(exception.ParseResponseError):
            resp_obj.extract_response(extract_binds_list)

        extract_binds_list = [
            {"resp_content_list_index_error": "content.person.cities.3"}
        ]
        resp_obj = response.ResponseObject(resp)

        with self.assertRaises(exception.ParseResponseError):
            resp_obj.extract_response(extract_binds_list)

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

        extract_binds_list = [
            {"resp_content_body": "content"}
        ]
        resp_obj = response.ResponseObject(resp)

        extract_binds_dict = resp_obj.extract_response(extract_binds_list)
        self.assertEqual(
            extract_binds_dict["resp_content_body"],
            "abc"
        )

    def test_extract_text_response(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'headers': {
                    'Content-Type': "application/json"
                },
                'body': "LB123abcRB789"
            }
        )

        extract_binds_list = [
            {"resp_content_key1": "LB123(.*)RB789"},
            {"resp_content_key2": "LB[\d]*(.*)RB[\d]*"},
            {"resp_content_key3": "LB[\d]*(.*)9"}
        ]
        resp_obj = response.ResponseObject(resp)

        extract_binds_dict = resp_obj.extract_response(extract_binds_list)
        self.assertEqual(
            extract_binds_dict["resp_content_key1"],
            "abc"
        )
        self.assertEqual(
            extract_binds_dict["resp_content_key2"],
            "abc"
        )
        self.assertEqual(
            extract_binds_dict["resp_content_key3"],
            "abcRB78"
        )

    def test_extract_text_response_exception(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'headers': {
                    'Content-Type': "application/json"
                },
                'body': "LB123abcRB789"
            }
        )
        extract_binds_list = [
            {"resp_content_key1": "LB123.*RB789"}
        ]
        resp_obj = response.ResponseObject(resp)
        with self.assertRaises(exception.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_empty(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'headers': {
                    'Content-Type': "application/json"
                },
                'body': ""
            }
        )

        extract_binds_list = [
            {"resp_content_body": "content"}
        ]
        resp_obj = response.ResponseObject(resp)
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)
        self.assertEqual(
            extract_binds_dict["resp_content_body"],
            ""
        )

        extract_binds_list = [
            {"resp_content_body": "content.abc"}
        ]
        resp_obj = response.ResponseObject(resp)
        with self.assertRaises(exception.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_validate(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)

        validators = [
            {"check": "resp_status_code", "comparator": "eq", "expect": 201},
            {"check": "resp_body_success", "comparator": "eq", "expect": True}
        ]
        variables_mapping = {
            "resp_status_code": 200,
            "resp_body_success": True
        }

        with self.assertRaises(exception.ValidationError):
            resp_obj.validate(validators, variables_mapping)

        validators = [
            {"check": "resp_status_code", "comparator": "eq", "expect": 201},
            {"check": "resp_body_success", "comparator": "eq", "expect": True}
        ]
        variables_mapping = {
            "resp_status_code": 201,
            "resp_body_success": True
        }

        self.assertTrue(resp_obj.validate(validators, variables_mapping))

    def test_validate_exception(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)

        # expected value missed in validators
        validators = [
            {"check": "status_code", "comparator": "eq", "expect": 201},
            {"check": "body_success", "comparator": "eq"}
        ]
        variables_mapping = {}
        with self.assertRaises(exception.ValidationError):
            resp_obj.validate(validators, variables_mapping)

        # expected value missed in variables mapping
        validators = [
            {"check": "resp_status_code", "comparator": "eq", "expect": 201},
            {"check": "body_success", "comparator": "eq"}
        ]
        variables_mapping = {
            "resp_status_code": 200
        }
        with self.assertRaises(exception.ValidationError):
            resp_obj.validate(validators, variables_mapping)
