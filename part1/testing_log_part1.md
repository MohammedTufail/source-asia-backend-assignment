PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Uri "http://localhost:8081/request" `
> >       -ContentType "application/json" `
> >       -Body '{"user_id":"user-1","payload":{"message":"hello"}}' |
> >     ConvertTo-Json -Depth 5
> >
> > }  
> > catch {  
> >  $\_.ErrorDetails.Message  
> > }  
> > {

    "status":  "accepted",
    "user_id":  "user-1",
    "message":  "201 created - Request accepted successfully."

}  
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> for ($i=1; $i

> >     try {
> >         Invoke-RestMethod `
> >           -Method POST `
> >           -Uri "http://localhost:8081/request" `
> >           -ContentType "application/json" `
> >           -Body '{"user_id":"user-limit","payload":{"count":1}}' |
> >         ConvertTo-Json -Depth 5
> >     }
> >     catch {
> >         $_.ErrorDetails.Message
> >     }
> >
> > }
> > {

    "status":  "accepted",
    "user_id":  "user-limit",
    "message":  "201 created - Request accepted successfully."

}
{
"status": "accepted",
"user_id": "user-limit",
"message": "201 created - Request accepted successfully."
{
"status": "accepted",
"user_id": "user-limit",
"message": "201 created - Request accepted successfully."
}
{
"status": "accepted",
"user_id": "user-limit",
"message": "201 created - Request accepted successfully."
}
{
"user_id": "user-limit",
"message": "201 created - Request accepted successfully."
}
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Method POST `
> >       -Uri "http://localhost:8081/request" `
> >       -ContentType "application/json" `
> >       -Body '{"user_id":"user-limit","payload":{"count":6}}' |
> >     ConvertTo-Json -Depth 5
> >
> > }
> > $\_.ErrorDetails.Message
> > }
> > {"error":"429 Too Many Requests - rate_limit_exceeded","message":"Rate limit exceeded: maximum 5 r

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Method POST `
> >       -Uri "http://localhost:8081/request" `
> >       -ContentType "application/json" `
> >       -Body '{"user_id":"user-1","payload":'
> >
> > }
> > $\_.ErrorDetails.Message
> > }
> > {"error":"400 Bad Request - invalid_json","message":"Request body must be valid JSON with user_id

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >       -Method POST `
> >       -Uri "http://localhost:8081/request" `
> >       -ContentType "application/json" `
> >       -Body '{"payload":{"hello":"world"}}'
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }
> > {"error":"400 Bad Request - missing_user_id","message":"user_id is required and must be a non-empt

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Method POST `
> >       -Uri "http://localhost:8081/request" `
> >       -ContentType "application/json" `
> >       -Body '{"user_id":"","payload":{"hello":"world"}}'
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }
> > {"error":"400 Bad Request - missing_user_id","message":"user_id is required and must be a non-empt

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> Invoke-RestMet

> > -Method GET `
> > -Uri "http://localhost:8081/stats" |
> > ConvertTo-Json -Depth 10
> > {

    "users":  {
                  "user-1":  {
                                 "rejected_cumulative":  0,
                                 "window_accepted":  0
                             },
                  "user-limit":  {
                                     "accepted":  5,
                                     "rejected_cumulative":  1,
                                     "window_accepted":  0
                                 }
              },
    "note":  "window_accepted reflects requests in the current 1-minute rolling window. rejected_c."

}
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> 1..20 | ForEac

> >     Start-Job {
> >         try {
> >             Invoke-RestMethod `
> >               -Method POST `
> >               -Uri "http://localhost:8081/request" `
> >               -ContentType "application/json" `
> >               -Body '{"user_id":"concurrent-user","payload":{"parallel":true}}'
> >         }
> >         catch {
> >             $_.ErrorDetails.Message
> >         }
> >     }
> >
> > }

Id Name PSJobTypeName State HasMoreData Location Command
1 Job1 BackgroundJob Running True localhost ...  
3 Job3 BackgroundJob Running True localhost ...  
5 Job5 BackgroundJob Running True localhost ...  
7 Job7 BackgroundJob Running True localhost ...  
9 Job9 BackgroundJob Running True localhost ...  
11 Job11 BackgroundJob Running True localhost ...  
13 Job13 BackgroundJob Running True localhost ...  
15 Job15 BackgroundJob Running True localhost ...  
17 Job17 BackgroundJob Running True localhost ...  
19 Job19 BackgroundJob Running True localhost ...  
21 Job21 BackgroundJob Running True localhost ...  
23 Job23 BackgroundJob Running True localhost ...  
25 Job25 BackgroundJob Running True localhost ...  
27 Job27 BackgroundJob Running True localhost ...  
29 Job29 BackgroundJob Running True localhost ...  
31 Job31 BackgroundJob Running True localhost ...  
33 Job33 BackgroundJob Running True localhost ...  
35 Job35 BackgroundJob Running True localhost ...  
37 Job37 BackgroundJob Running True localhost ...  
39 Job39 BackgroundJob Running True localhost ...

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> Invoke-RestMet

> > -Method GET `
> > -Uri "http://localhost:8081/stats" |
> > ConvertTo-Json -Depth 10
> > {

    "users":  {
                  "concurrent-user":  {
                                          "accepted":  5,
                                          "rejected_cumulative":  15,
                                          "window_accepted":  5
                                      },
                  "user-1":  {
                                 "accepted":  1,
                                 "rejected_cumulative":  0,
                                 "window_accepted":  0
                             },
                  "user-limit":  {
                                     "accepted":  5,
                                     "rejected_cumulative":  1,
                                     "window_accepted":  0
                                 }
              },
    "note":  "window_accepted reflects requests in the current 1-minute rolling window. rejected_c."

}
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> go test ./part
=== RUN TestRateLimiter_AllowsUpToFive
--- PASS: TestRateLimiter_AllowsUpToFive (0.00s)
--- PASS: TestRateLimiter_IndependentUsers (0.00s)
=== RUN TestRateLimiter_ConcurrentSafety
--- PASS: TestRateLimiter_ConcurrentSafety (0.00s)
=== RUN TestRateLimiter_StatsRejectedCumulative
--- PASS: TestRateLimiter_StatsRejectedCumulative (0.00s)
=== RUN TestHandler_AcceptsValidRequest
--- PASS: TestHandler_AcceptsValidRequest (0.00s)
=== RUN TestHandler_RejectsMissingUserID
--- PASS: TestHandler_RejectsMissingUserID (0.00s)
=== RUN TestHandler_RejectsEmptyUserID
--- PASS: TestHandler_RejectsEmptyUserID (0.00s)
=== RUN TestHandler_RejectsMissingPayload
--- PASS: TestHandler_RejectsMissingPayload (0.00s)
=== RUN TestHandler_RejectsInvalidJSON
--- PASS: TestHandler_RejectsInvalidJSON (0.00s)
=== RUN TestHandler_RateLimit
--- PASS: TestHandler_RateLimit (0.00s)
=== RUN TestHandler_StatsEndpoint
--- PASS: TestHandler_StatsEndpoint (0.00s)
=== RUN TestHandler_ConcurrentRequests
--- PASS: TestHandler_ConcurrentRequests (0.00s)
=== RUN TestRateLimiter_WindowExpiry
--- PASS: TestRateLimiter_WindowExpiry (0.00s)
PASS
ok source-asia-backend-assignment/part1 (cached)
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> go test -race
=== RUN TestRateLimiter_AllowsUpToFive
--- PASS: TestRateLimiter_AllowsUpToFive (0.00s)
--- PASS: TestRateLimiter_IndependentUsers (0.00s)
=== RUN TestRateLimiter_ConcurrentSafety
--- PASS: TestRateLimiter_ConcurrentSafety (0.00s)
=== RUN TestRateLimiter_StatsRejectedCumulative
--- PASS: TestRateLimiter_StatsRejectedCumulative (0.00s)
=== RUN TestHandler_AcceptsValidRequest
=== RUN TestHandler_RejectsMissingUserID
--- PASS: TestHandler_RejectsMissingUserID (0.00s)
=== RUN TestHandler_RejectsEmptyUserID
--- PASS: TestHandler_RejectsEmptyUserID (0.00s)
=== RUN TestHandler_RejectsMissingPayload
--- PASS: TestHandler_RejectsMissingPayload (0.00s)
--- PASS: TestHandler_RejectsInvalidJSON (0.00s)
=== RUN TestHandler_RateLimit
--- PASS: TestHandler_RateLimit (0.01s)
=== RUN TestHandler_StatsEndpoint
--- PASS: TestHandler_StatsEndpoint (0.01s)
=== RUN TestHandler_ConcurrentRequests
--- PASS: TestHandler_ConcurrentRequests (0.03s)
=== RUN TestRateLimiter_WindowExpiry
--- PASS: TestRateLimiter_WindowExpiry (0.00s)
PASS
ok source-asia-backend-assignment/part1 1.633s
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> go test ./part
=== RUN TestRateLimiter_ConcurrentSafety
--- PASS: TestRateLimiter_ConcurrentSafety (0.00s)
=== RUN TestHandler_ConcurrentRequests
--- PASS: TestHandler_ConcurrentRequests (0.01s)
PASS
ok source-asia-backend-assignment/part1 0.485s
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> go test ./part
testing: warning: no tests to run
PASS
ok source-asia-backend-assignment/part1 0.410s [no tests to run]
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> Invoke-RestMet

> > -Method GET `
> > -Uri "http://localhost:8081/stats" |
> > ConvertTo-Json -Depth 10
> > {

    "users":  {
                  "concurrent-user":  {
                                          "accepted":  5,
                                          "rejected_cumulative":  15,
                                          "window_accepted":  0
                                      },
                  "user-1":  {
                                 "accepted":  1,
                                 "rejected_cumulative":  0,
                                 "window_accepted":  0
                             },
                  "user-limit":  {
                                     "accepted":  5,
                                     "rejected_cumulative":  1,
                                     "window_accepted":  0
                                 }
              },
    "note":  "window_accepted reflects requests in the current 1-minute rolling window. rejected_c."

}
