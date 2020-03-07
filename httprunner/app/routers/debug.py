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

    project_meta_json = project_meta.dict(by_alias=True)
    if project_meta.debugtalk_py:
        origin_local_keys = list(locals().keys()).copy()
        exec(project_meta.debugtalk_py, {}, locals())
        new_local_keys = list(locals().keys()).copy()
        new_added_keys = set(new_local_keys) - set(origin_local_keys)
        new_added_keys.remove("origin_local_keys")
        project_meta_json["functions"] = {}
        for func_name in new_added_keys:
            project_meta_json["functions"][func_name] = locals()[func_name]

    testcase_json = testcase.dict(by_alias=True)
    tests_mapping = {
        "project_mapping": project_meta_json,
        "testcases": [testcase_json]
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
