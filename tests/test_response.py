import requests
from httprunner import exceptions, response, utils
from httprunner.compat import bytes, str
from tests.base import ApiServerUnittest


class TestResponse(ApiServerUnittest):

    def setUp(self):
        imported_module = utils.get_imported_module("httprunner.built_in")
        self.functions_mapping = utils.filter_module(imported_module, "function")

    def test_parse_response_object_json(self):
        url = "http://127.0.0.1:5000/api/users"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)
        self.assertTrue(hasattr(resp_obj, 'status_code'))
        self.assertTrue(hasattr(resp_obj, 'headers'))
        self.assertTrue(hasattr(resp_obj, 'content'))
        self.assertIn('Content-Type', resp_obj.headers)
        self.assertIn('Content-Length', resp_obj.headers)
        self.assertIn('success', resp_obj.json)

    def test_parse_response_object_content(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)
        self.assertEqual(bytes, type(resp_obj.content))

    def test_extract_response_status_code(self):
        resp = requests.get(url="http://127.0.0.1:3458/status/200")
        resp_obj = response.ResponseObject(resp)

        extract_binds_list = [
            {"resp_status_code": "status_code"}
        ]
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)

        self.assertEqual(
            extract_binds_dict["resp_status_code"],
            200
        )

        extract_binds_list = [
            {"resp_status_code": "status_code.xx"}
        ]
        with self.assertRaises(exceptions.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_cookies(self):
        resp = requests.get(
            url="http://127.0.0.1:3458/cookies",
            headers={
                "accept": "application/json"
            }
        )
        resp_obj = response.ResponseObject(resp)

        extract_binds_list = [
            {"resp_cookies": "cookies"}
        ]
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)
        self.assertEqual(
            extract_binds_dict["resp_cookies"],
            {}
        )

        extract_binds_list = [
            {"resp_cookies": "cookies.xx"}
        ]
        with self.assertRaises(exceptions.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_elapsed(self):
        resp = requests.post(
            url="http://127.0.0.1:3458/anything",
            json={
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
        )
        resp_obj = response.ResponseObject(resp)

        extract_binds_list = [
            {"resp_elapsed": "elapsed"}
        ]
        with self.assertRaises(exceptions.ParamsError):
            resp_obj.extract_response(extract_binds_list)

        extract_binds_list = [
            {"resp_elapsed_microseconds": "elapsed.microseconds"},
            {"resp_elapsed_seconds": "elapsed.seconds"},
            {"resp_elapsed_days": "elapsed.days"},
            {"resp_elapsed_total_seconds": "elapsed.total_seconds"}
        ]
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)
        self.assertGreater(extract_binds_dict["resp_elapsed_microseconds"], 1000)
        self.assertEqual(extract_binds_dict["resp_elapsed_seconds"], 0)
        self.assertEqual(extract_binds_dict["resp_elapsed_days"], 0)
        self.assertGreater(extract_binds_dict["resp_elapsed_total_seconds"], 0)

        extract_binds_list = [
            {"resp_elapsed": "elapsed.years"}
        ]
        with self.assertRaises(exceptions.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_headers(self):
        resp = requests.get(url="http://127.0.0.1:3458/status/200")
        resp_obj = response.ResponseObject(resp)

        extract_binds_list = [
            {"resp_headers": "headers"},
            {"resp_headers_content_type": "headers.Content-Type"},
            {"resp_headers_content_type_lowercase": "headers.content-type"}
        ]
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)
        self.assertIn("Content-Type", extract_binds_dict["resp_headers"])
        self.assertIn("text/html", extract_binds_dict["resp_headers_content_type"])
        self.assertIn("text/html", extract_binds_dict["resp_headers_content_type_lowercase"])

        extract_binds_list = [
            {"resp_headers_xxx": "headers.xxx"}
        ]
        with self.assertRaises(exceptions.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_body_json(self):
        resp = requests.post(
            url="http://127.0.0.1:3458/anything",
            json={
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
        )
        # resp.json()
        # {
        #     "args": {},
        #     "data": "{\"success\": false, \"person\": {\"name\": {\"first_name\": \"Leo\", \"last_name\": \"Lee\"}, \"age\": 29, \"cities\": [\"Guangzhou\", \"Shenzhen\"]}}",
        #     "files": {},
        #     "form": {},
        #     "headers": {
        #         "Accept": "*/*",
        #         "Accept-Encoding": "gzip, deflate",
        #         "Connection": "keep-alive",
        #         "Content-Length": "129",
        #         "Content-Type": "application/json",
        #         "Host": "127.0.0.1:3458",
        #         "User-Agent": "python-requests/2.18.4"
        #     },
        #     "json": {
        #         "person": {
        #         "age": 29,
        #         "cities": [
        #             "Guangzhou",
        #             "Shenzhen"
        #         ],
        #         "name": {
        #             "first_name": "Leo",
        #             "last_name": "Lee"
        #         }
        #         },
        #         "success": false
        #     },
        #     "method": "POST",
        #     "origin": "127.0.0.1",
        #     "url": "http://127.0.0.1:3458/anything"
        # }

        extract_binds_list = [
            {"resp_headers_content_type": "headers.content-type"},
            {"resp_content_body_success": "json.json.success"},
            {"resp_content_content_success": "content.json.success"},
            {"resp_content_text_success": "text.json.success"},
            {"resp_content_person_first_name": "content.json.person.name.first_name"},
            {"resp_content_cities_1": "content.json.person.cities.1"}
        ]
        resp_obj = response.ResponseObject(resp)
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)

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

    def test_extract_response_body_html(self):
        resp = requests.get(url="http://127.0.0.1:3458/")
        resp_obj = response.ResponseObject(resp)

        extract_binds_list = [
            {"resp_content": "content"}
        ]
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)

        self.assertIsInstance(extract_binds_dict["resp_content"], str)
        self.assertIn("python-requests.org", extract_binds_dict["resp_content"])

        extract_binds_list = [
            {"resp_content": "content.xxx"}
        ]
        with self.assertRaises(exceptions.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_others(self):
        resp = requests.get(url="http://127.0.0.1:3458/status/200")
        resp_obj = response.ResponseObject(resp)

        extract_binds_list = [
            {"resp_others_encoding": "encoding"},
            {"resp_others_history": "history"}
        ]
        with self.assertRaises(exceptions.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_fail(self):
        resp = requests.post(
            url="http://127.0.0.1:3458/anything",
            json={
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
        )

        extract_binds_list = [
            {"resp_content_dict_key_error": "content.not_exist"}
        ]
        resp_obj = response.ResponseObject(resp)

        with self.assertRaises(exceptions.ParseResponseFailure):
            resp_obj.extract_response(extract_binds_list)

        extract_binds_list = [
            {"resp_content_list_index_error": "content.person.cities.3"}
        ]
        resp_obj = response.ResponseObject(resp)

        with self.assertRaises(exceptions.ParseResponseFailure):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_json_string(self):
        resp = requests.post(
            url="http://127.0.0.1:3458/anything",
            data="abc"
        )

        extract_binds_list = [
            {"resp_content_body": "content.data"}
        ]
        resp_obj = response.ResponseObject(resp)

        extract_binds_dict = resp_obj.extract_response(extract_binds_list)
        self.assertEqual(
            extract_binds_dict["resp_content_body"],
            "abc"
        )

    def test_extract_text_response(self):
        resp = requests.post(
            url="http://127.0.0.1:3458/anything",
            data="LB123abcRB789"
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
            url="http://127.0.0.1:3458/anything",
            data="LB123abcRB789"
        )
        extract_binds_list = [
            {"resp_content_key1": "LB123.*RB789"}
        ]
        resp_obj = response.ResponseObject(resp)
        with self.assertRaises(exceptions.ParamsError):
            resp_obj.extract_response(extract_binds_list)

    def test_extract_response_empty(self):
        resp = requests.post(
            url="http://127.0.0.1:3458/anything",
            data="abc"
        )

        extract_binds_list = [
            {"resp_content_body": "content.data"}
        ]
        resp_obj = response.ResponseObject(resp)
        extract_binds_dict = resp_obj.extract_response(extract_binds_list)
        self.assertEqual(
            extract_binds_dict["resp_content_body"],
            'abc'
        )

        extract_binds_list = [
            {"resp_content_body": "content.data.def"}
        ]
        resp_obj = response.ResponseObject(resp)
        with self.assertRaises(exceptions.ParseResponseFailure):
            resp_obj.extract_response(extract_binds_list)
