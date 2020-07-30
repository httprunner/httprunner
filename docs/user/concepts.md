
## variables priority

There are several different types of `variables`, and the priory can be confusing. The best way to avoid confusion is to use different variable names. However, if you must use the same variable names, you should understand the priority strategy.

### testcase

```yaml
config:
    name: xxx
    variables:              # config variables
        varA: "configA"
        varB: "configB"
        varC: "configC"
    parameters:             # parameter variables
        varA: ["paramA1"]
        varB: ["paramB1"]

teststeps:
-
    name: step 1
    variables:              # step variables
        varA: "step1A"
    request:
        url: /$varA/$varB/$varC # varA="step1A", varB="paramB1", varC="configC"
        method: GET
    extract:                # extracted variables
        varA: body.data.A   # suppose varA="extractVarA"
        varB: body.data.B   # suppose varB="extractVarB"
-
    name: step 2
    varialbes:
        varA: "step2A"
    request:
        url: /$varA/$varB/$varC # varA="step2A", varB="extractVarB", varC="configC"
        method: GET
```

In a testcase, variables priority are in the following order:

- step variables > extracted variables, e.g. step 2, varA="step2A"
- parameter variables > config variables, e.g. step 1, varB="paramB1"
- extracted variables > parameter variables > config variables, e.g. step 2, varB="extractVarB"
- config variables are in the lowest priority, e.g. step 1/2, varC="configC"

### testsuite

```yaml
config:
    name: xxx
    variables:                  # testsuite config variables
        varA: "configA"
        varB: "configB"
        varC: "configC"

testcases:
-
    name: case 1
    variables:                  # testcase variables
        varA: "case1A"
    testcase: /path/to/testcase1
    export: ["varA", "varB"]    # export variables
-
    name: case 2
    varialbes:                  # testcase variables
        varA: "case2A"
    testcase: /path/to/testcase2
```

In a testsuite, variables priority are in the following order:

- testcase variables > export variables > testsuite config variables > referenced testcase config variables
