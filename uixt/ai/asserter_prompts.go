package ai

// Default assertion system prompt
const defaultAssertionPrompt = `You are a senior testing engineer. User will give an assertion and a screenshot of a page. By carefully viewing the screenshot, please tell whether the assertion is truthy.`

// Default assertion response format
const defaultAssertionResponseJsonFormat = `Return in the following JSON format:
{
  pass: boolean, // whether the assertion is truthy
  thought: string | null, // string, if the result is falsy, give the reason why it is falsy. Otherwise, put null.
}`

// UI-TARS assertion response format
const uiTarsAssertionResponseFormat = `## Output Json String Format
` + "```" + `
"{
  "pass": <<is a boolean value from the enum [true, false], true means the assertion is truthy>>,
  "thought": "<<is a string, give the reason why the assertion is falsy or truthy. Otherwise.>>"
}"
` + "```" + `

## Rules **MUST** follow
- Make sure to return **only** the JSON, with **no additional** text or explanations.
- Use Chinese in ` + "`Thought`" + ` part.
- You **MUST** strictly follow up the **Output Json String Format**.`
