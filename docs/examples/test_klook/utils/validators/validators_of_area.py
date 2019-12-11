def check_search_area_result(content, expect_name):
    print(content, expect_name)
    found = False
    for item in content:
        if item['fullName'] == expect_name:
            found = True
            break
    assert found


def exists_default_group(content, expect):
    found = False
    for item in content:
        if item['defaultGroup']:
            print('defaultGroup found, id={}, parentAreaId={}'.format(item['id'], item['parentAreaId']))
            found = True
            break
    assert found
