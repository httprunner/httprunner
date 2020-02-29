from fastapi import APIRouter

from httprunner.api import HttpRunner
from httprunner.schema import ProjectMeta, TestCase

router = APIRouter()
runner = HttpRunner()


@router.post("/hrun/debug/testcase", tags=["debug"])
async def debug_single_testcase(project_meta: ProjectMeta, testcase: TestCase):
    resp = {
        "code": 0,
        "message": "success",
        "result": {}
    }
    tests_mapping = {
        "project_mapping": project_meta,
        "testcases": [testcase]
    }
    summary = runner.run_tests(tests_mapping)
    if not summary["success"]:
        resp["code"] = 1
        resp["message"] = "fail"

    resp["result"] = summary
    return resp


# @router.post("/hrun/debug/api", tags=["debug"])
# async def debug_single_api():
#     resp = {
#         "code": 0,
#         "message": "success",
#         "result": {}
#     }
#
#     # tests_mapping
#
#     # summary = runner.run_tests(tests_mapping)
#
#     return resp
#
#
# @router.post("/hrun/debug/testcases", tags=["debug"])
# async def debug_multiple_testcases(project_meta: ProjectMeta, testcases: TestCases):
#     tests_mapping = {
#         "project_mapping": project_meta,
#         "testcases": testcases
#     }
