from fastapi import APIRouter

from httprunner.api import HttpRunner

router = APIRouter()
runner = HttpRunner()


@router.get("/hrun/debug/api", tags=["testcase"])
async def debug_single_api():
    pass


@router.get("/hrun/debug/testcase", tags=["testcase"])
async def debug_single_testcase():
    resp = {
        "code": 0,
        "message": "success",
        "result": {}
    }
    testcases = [
        {
            "config": {
                'name': "post data",
                'variables': {
                    "var1": "abc",
                    "var2": "def"
                },
                "export": ["status_code", "req_data"]
            },
            "teststeps": [
                {
                    "name": "post data",
                    "request": {
                        "url": "http://httpbin.org/post",
                        "method": "POST",
                        "headers": {
                            "User-Agent": "python-requests/2.18.4",
                            "Content-Type": "application/json"
                        },
                        "data": "$var1"
                    },
                    "extract": {
                        "status_code": "status_code",
                        "req_data": "content.data"
                    },
                    "validate": [
                        {"eq": ["status_code", 201]}
                    ]
                }
            ]
        }
    ]
    tests_mapping = {
        "testcases": testcases
    }
    summary = runner.run_tests(tests_mapping)
    if not summary["success"]:
        resp["code"] = 1
        resp["message"] = "fail"

    resp["result"] = summary
    return resp
