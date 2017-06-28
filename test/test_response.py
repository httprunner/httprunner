import random
import requests
from ate import response, context, exception
from test.base import ApiServerUnittest

class TestResponse(ApiServerUnittest):

    def test_parse_response_object_json(self):
        url = "http://127.0.0.1:5000/api/users"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)
        parse_result = resp_obj.parse_response_object()
        self.assertIn('status_code', parse_result)
        self.assertIn('headers', parse_result)
        self.assertIn('body', parse_result)
        self.assertIn('Content-Type', parse_result['headers'])
        self.assertIn('Content-Length', parse_result['headers'])
        self.assertIn('success', parse_result['body'])

    def test_parse_response_object_text(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)
        parse_result = resp_obj.parse_response_object()
        self.assertIn('status_code', parse_result)
        self.assertIn('headers', parse_result)
        self.assertIn('body', parse_result)
        self.assertIn('Content-Type', parse_result['headers'])
        self.assertIn('Content-Length', parse_result['headers'])
        self.assertTrue(str, type(parse_result['body']))

    def test_diff_response_status_code_equal(self):
        status_code = random.randint(200, 511)
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'status_code': status_code,
            }
        )

        expected_resp_json = {
            'status_code': status_code
        }
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertFalse(diff_content)

    def test_diff_response_status_code_not_equal(self):
        status_code = random.randint(200, 511)
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'status_code': status_code,
            }
        )

        expected_resp_json = {
            'status_code': 512
        }
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertIn('value', diff_content['status_code'])
        self.assertIn('expected', diff_content['status_code'])
        self.assertEqual(diff_content['status_code']['value'], status_code)
        self.assertEqual(diff_content['status_code']['expected'], 512)

    def test_diff_response_headers_equal(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'headers': {
                    'abc': 123,
                    'def': 456
                }
            }
        )

        expected_resp_json = {
            'headers': {
                'abc': 123,
                'def': '456'
            }
        }
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertFalse(diff_content)

    def test_diff_response_headers_not_equal(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'headers': {
                    'a': 123,
                    'b': '456',
                    'c': '789'
                }
            }
        )

        expected_resp_json = {
            'headers': {
                'a': '123',
                'b': '457',
                'd': 890
            }
        }
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertEqual(
            diff_content['headers'],
            {
                'b': {'expected': '457', 'value': '456'},
                'd': {'expected': 890, 'value': None}
            }
        )

    def test_diff_response_body_equal(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'body': {
                    'success': True,
                    'count': 10
                }
            }
        )

        # expected response body is not specified
        expected_resp_json = {}
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertFalse(diff_content)

        # response body is the same as expected response body
        expected_resp_json = {
            'body': {
                'success': True,
                'count': '10'
            }
        }
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertFalse(diff_content)

    def test_diff_response_body_not_equal_type_unmatch(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'body': {
                    'success': True,
                    'count': 10
                }
            }
        )

        # response body content type not match
        expected_resp_json = {
            'body': "ok"
        }
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertEqual(
            diff_content['body'],
            {
                'value': {'success': True, 'count': 10},
                'expected': 'ok'
            }
        )

    def test_diff_response_body_not_equal_string_unmatch(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'body': "success"
            }
        )

        # response body content type matched to be string, while value unmatch
        expected_resp_json = {
            'body': "ok"
        }
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertEqual(
            diff_content['body'],
            {
                'value': 'success',
                'expected': 'ok'
            }
        )

    def test_diff_response_body_not_equal_json_unmatch(self):
        resp = requests.post(
            url="http://127.0.0.1:5000/customize-response",
            json={
                'body': {
                    'success': False
                }
            }
        )

        # response body is the same as expected response body
        expected_resp_json = {
            'body': {
                'success': True,
                'count': 10
            }
        }
        resp_obj = response.ResponseObject(resp)
        diff_content = resp_obj.diff_response(expected_resp_json)
        self.assertEqual(
            diff_content['body'],
            {
                'success': {
                    'value': False,
                    'expected': True
                },
                'count': {
                    'value': None,
                    'expected': 10
                }
            }
        )

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

        test_context = context.Context()
        test_context.bind_extractors(extract_binds)
        resp_obj = response.ResponseObject(resp)
        resp_obj.extract_response(test_context)

        extract_binds_dict = test_context.extractors
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

        test_context = context.Context()
        test_context.bind_extractors(extract_binds)
        resp_obj = response.ResponseObject(resp)

        with self.assertRaises(exception.ParamsError):
            resp_obj.extract_response(test_context)

        extract_binds = {
            "resp_content_list_index_error": "content.person.cities.3"
        }

        test_context = context.Context()
        test_context.bind_extractors(extract_binds)
        resp_obj = response.ResponseObject(resp)

        with self.assertRaises(exception.ParamsError):
            resp_obj.extract_response(test_context)

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

        test_context = context.Context()
        test_context.bind_extractors(extract_binds)
        resp_obj = response.ResponseObject(resp)
        resp_obj.extract_response(test_context)

        extract_binds_dict = test_context.extractors
        self.assertEqual(
            extract_binds_dict["resp_content_body"],
            "abc"
        )
