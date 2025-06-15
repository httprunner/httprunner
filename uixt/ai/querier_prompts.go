package ai

// Default query system prompt
const defaultQueryPrompt = `You are an AI assistant specialized in analyzing images and extracting information. User will provide a screenshot and a query asking for specific information to be extracted from the image. Please analyze the image carefully and provide the requested information.`

// UI-TARS query response format
const uiTarsQueryResponseFormat = `
## Output Json String Format
` + "```" + `
"{
  "content": "<<is a string containing the extracted information or analysis result>>",
  "thought": "<<is a string explaining your analysis process and reasoning. Use Chinese.>>"
}"
` + "```" + `

## Rules **MUST** follow
- Make sure to return **only** the JSON, with **no additional** text or explanations.
- Use Chinese in ` + "`Thought`" + ` part.
- You **MUST** strictly follow up the **Output Json String Format**.
- Provide detailed and accurate information extraction based on the image content.`
