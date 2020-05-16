import logging
import subprocess
from typing import List

import pkg_resources
from fastapi import APIRouter

router = APIRouter()


@router.get("/hrun/deps", tags=["deps"])
async def get_installed_dependenies():
    resp = {"code": 0, "message": "success", "result": {}}
    for p in pkg_resources.working_set:
        resp["result"][p.project_name] = p.version

    return resp


@router.post("/hrun/deps", tags=["deps"])
async def install_dependenies(deps: List[str]):
    resp = {"code": 0, "message": "success", "result": {}}
    for dep in deps:
        try:
            p = subprocess.run(["pip", "install", dep])
            assert p.returncode == 0
            resp["result"][dep] = True
        except (AssertionError, subprocess.SubprocessError):
            resp["result"][dep] = False
            resp["code"] = 1
            resp["message"] = "fail"
            logging.error(f"failed to install dependency: {dep}")

    return resp
