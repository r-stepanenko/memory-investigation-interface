# SEARCH

$response = Invoke-RestMethod `
    -Uri "http://localhost:8080/api/search" `
    -Method POST `
    -ContentType "application/json" `
    -Body '
{
  "dataset_id":"control",
  "hints":{
    "user_id":"ivan"
  },
  "context":{
    "before":"30m",
    "after":"30m",
    "require_nearby":[
      {
        "action":"file_copy"
      }
    ]
  }
}
'

$id = $response.search_id

Write-Host "SEARCH ID:" $id


# RESULT

Invoke-RestMethod `
    -Uri "http://localhost:8080/api/search/$id" `
    -Method GET


# CONTEXT

Invoke-RestMethod `
    -Uri "http://localhost:8080/api/events/evt_32/context" `
    -Method GET


# EXPLAIN

$explainUrl = "http://localhost:8080/api/search/$($id)/candidates/evt_32/explain"

Write-Host $explainUrl

Invoke-RestMethod `
    -Uri $explainUrl `
    -Method GET