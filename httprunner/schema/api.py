from pydantic import BaseModel

from httprunner.schema import common


class Api(BaseModel):
    name: common.Name
    request: common.Request
    variables: common.Variables
    base_url: common.BaseUrl
    setup_hooks: common.Hook
    teardown_hooks: common.Hook
    extract: common.Extract
    validate: common.Validate
