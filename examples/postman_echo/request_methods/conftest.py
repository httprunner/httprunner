from typing import List

import pytest
from loguru import logger

from httprunner.schema import TConfig, TStep


@pytest.fixture(scope="session", autouse=True)
def session_fixture(request):
    """setup and teardown each task"""
    total_testcases_num = request.node.testscollected
    testcases = []
    for item in request.node.items:
        testcase = {
            "name": item.cls.config.name,
            "path": item.cls.config.path,
            "node_id": item.nodeid,
        }
        testcases.append(testcase)

    logger.debug(f"collected {total_testcases_num} testcases: {testcases}")

    yield

    logger.debug(f"teardown task fixture")

    # teardown task
    # TODO: upload task summary


@pytest.fixture(scope="function", autouse=True)
def testcase_fixture(request):
    """setup and teardown each testcase"""
    config: TConfig = request.cls.config
    teststeps: List[TStep] = request.cls.teststeps

    logger.debug(f"setup testcase fixture: {config.name} - {request.module.__name__}")

    # prefix = f"HRUN-{uuid.uuid4()}"
    # for index, teststep in enumerate(teststeps):
    #     # you can update testcase teststep here
    #     teststep.request.headers["HRUN-Request-ID"] = f"{prefix}-{index}"

    yield

    logger.debug(f"teardown testcase fixture: {config.name} - {request.module.__name__}")

    summary = request.instance.get_summary()
    logger.debug(f"testcase result summary: {summary}")

    # TODO: upload testcase summary
