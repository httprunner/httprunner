def klook_len_eq(check_value, expect_value):
    # 校验列表、字典、字符串等长度是否相等
    assert len(check_value) == int(expect_value)
