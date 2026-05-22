PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Method POST `
> >       -Uri "http://localhost:8082/products" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "name":"Widget A",
> >         "sku":"SKU-001",
> >         "image_urls":[
> >           "https://cdn.example.com/products/sku-001/img1.jpg"
> >         ],
> >         "video_urls":[
> >           "https://cdn.example.com/products/sku-001/demo.mp4"
> >         ]
> >       }' |
> >     ConvertTo-Json -Depth 10
> >
> > }  
> > catch {
> > }  
> > {

    "id":  "59a75a00-731d-90e2-b5af-4eaf2e3d5fd6",
    "name":  "Widget A",
    "sku":  "SKU-001",
    "image_urls":  [
                       "https://cdn.example.com/products/sku-001/img1.jpg"
                   ],
    "video_urls":  [
                       "https://cdn.example.com/products/sku-001/demo.mp4"
                   ],
    "created_at":  "2026-05-22T22:51:35.2795853Z"

}
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >       -Method POST `
> >       -Uri "http://localhost:8082/products" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "name":"Duplicate Widget",
> >         "sku":"SKU-001"
> >       }'
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }
> > {"error":"duplicate_sku","message":"product with SKU \"SKU-001\" already exists"}

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >       -Method POST `
> >       -Uri "http://localhost:8082/products" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "name":"",
> >         "sku":"SKU-EMPTY"
> >       }'
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }
> > {"error":"validation_error","message":"name is required and must be a non-empty string"}

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >       -Method POST `
> >       -Uri "http://localhost:8082/products" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "name":"Widget",
> >         "sku":""
> >       }'
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }
> > {"error":"validation_error","message":"sku is required and must be a non-empty string"}

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Uri "http://localhost:8082/products" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "name":"Widget",
> >         "sku":""
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }
> > {"error":"validation_error","message":"sku is required and must be a non-empty string"}

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Method POST `
> >       -Uri "http://localhost:8082/products" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "name":"Widget",
> >         "sku":"SKU-BADURL",
> >         "image_urls":["not-a-url"]
> >       }'
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> Invoke-RestMet

> > -Method GET `
> > -Uri "http://localhost:8082/products" |
> > ConvertTo-Json -Depth 10
> > {

    "data":  [
                 {
                     "id":  "59a75a00-731d-90e2-b5af-4eaf2e3d5fd6",
                     "name":  "Widget A",
                     "sku":  "SKU-001",
                     "image_count":  1,
                     "video_count":  1,
                     "thumbnail_url":  "https://cdn.example.com/products/sku-001/img1.jpg",
                     "created_at":  "2026-05-22T22:51:35.2795853Z"
                 }
             ],
    "total":  1,
    "limit":  20,
    "offset":  0,

}
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> Invoke-RestMet

> > -Method GET `
> > -Uri "http://localhost:8082/products?limit=3&offset=0" |
> > ConvertTo-Json -Depth 10
> > {

    "data":  [
                 {
                     "id":  "59a75a00-731d-90e2-b5af-4eaf2e3d5fd6",
                     "name":  "Widget A",
                     "sku":  "SKU-001",
                     "image_count":  1,
                     "video_count":  1,
                     "created_at":  "2026-05-22T22:51:35.2795853Z"
                 }
             ],
    "total":  1,
    "limit":  3,
    "offset":  0,
    "has_more":  false

}
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> Invoke-RestMet

> > -Method GET `
> > -Uri "http://localhost:8082/products/YOUR_ID" |
> > ConvertTo-Json -Depth 10
> > Invoke-RestMethod : {"error":"not_found","message":"No product found with the given ID."}
> > At line:1 char:1

- Invoke-RestMethod `
- ```
      + CategoryInfo          : InvalidOperation: (System.Net.HttpWebRequest:HttpWebRequest) [Invoke
      + FullyQualifiedErrorId : WebCmdletWebResponseException,Microsoft.PowerShell.Commands.InvokeRe
  PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {
  >>       -Method GET `
  >>       -Uri "http://localhost:8082/products/unknown-id"
  >> }
  >> catch {
  >>     $_.ErrorDetails.Message
  >> }
  {"error":"not_found","message":"No product found with the given ID."}
  ```

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Method POST `
> >       -Uri "http://localhost:8082/products/59a75a00-731d-90e2-b5af-4eaf2e3d5fd6/media" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "image_urls":[
> >           "https://cdn.example.com/new-image.jpg"
> >         ],
> >         "video_urls":[
> >           "https://cdn.example.com/new-video.mp4"
> >         ]
> >       }' |
> >     ConvertTo-Json -Depth 10
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }

    "id":  "59a75a00-731d-90e2-b5af-4eaf2e3d5fd6",
    "name":  "Widget A",
    "sku":  "SKU-001",
    "image_urls":  [
                       "https://cdn.example.com/products/sku-001/img1.jpg",
                       "https://cdn.example.com/new-image.jpg"
                   ],
    "video_urls":  [
                       "https://cdn.example.com/products/sku-001/demo.mp4",
                       "https://cdn.example.com/new-video.mp4"
                   ],
    "created_at":  "2026-05-22T22:51:35.2795853Z"

}
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Uri "http://localhost:8082/products/YOUR_ID/media" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "image_urls":[],
> >         "video_urls":[]
> >       }'
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }
> > {"error":"not_found","message":"No product found with the given ID."}

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> try {

> >     Invoke-RestMethod `
> >       -Method POST `
> >       -Uri "http://localhost:8082/products/unknown-id/media" `
> >       -ContentType "application/json" `
> >       -Body '{
> >         "image_urls":[
> >           "https://cdn.example.com/test.jpg"
> >         ]
> >       }'
> >
> > }
> > catch {
> > $\_.ErrorDetails.Message
> > }
> > {"error":"not_found","message":"No product found with the given ID."}

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> for ($i=1; $i

> >         name = "Product-$i"
> >         sku  = "PERF-SKU-$i"
> >
> >         image_urls = @(
> >             "https://cdn.example.com/$i/img1.jpg",
> >             "https://cdn.example.com/$i/img2.jpg",
> >             "https://cdn.example.com/$i/img4.jpg",
> >             "https://cdn.example.com/$i/img5.jpg",
> >             "https://cdn.example.com/$i/img6.jpg",
> >             "https://cdn.example.com/$i/img7.jpg",
> >             "https://cdn.example.com/$i/img8.jpg",
> >             "https://cdn.example.com/$i/img9.jpg",
> >             "https://cdn.example.com/$i/img10.jpg"
> >         )
> >
> >         video_urls = @()
> >
> >     } | ConvertTo-Json -Depth 5
> >
> >     Invoke-RestMethod `
> >       -Method POST `
> >       -Uri "http://localhost:8082/products" `
> >       -ContentType "application/json" `
> >       -Body $body | Out-Null
> >
> > }
> > PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> Measure-Comman
> > Invoke-RestMethod `      -Method GET`
> > -Uri "http://localhost:8082/products?limit=20"
> >
> > }

Days : 0
Hours : 0
Seconds : 0
Milliseconds : 1
Ticks : 19525
TotalDays : 2.25983796296296E-08
TotalHours : 5.42361111111111E-07
TotalMinutes : 3.25416666666667E-05
TotalSeconds : 0.0019525
TotalMilliseconds : 1.9525

PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> $response = In

> > -Method GET `
> > -Uri "http://localhost:8082/products?limit=1"
> >
> > $response.data | ConvertTo-Json -Depth 10
> > {

    "id":  "59a75a00-731d-90e2-b5af-4eaf2e3d5fd6",
    "name":  "Widget A",
    "sku":  "SKU-001",
    "image_count":  2,
    "video_count":  2,
    "thumbnail_url":  "https://cdn.example.com/products/sku-001/img1.jpg",
    "created_at":  "2026-05-22T22:51:35.2795853Z"

}
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> go test ./part
=== RUN TestValidator_AcceptsHTTPS
--- PASS: TestValidator_AcceptsHTTPS (0.00s)
=== RUN TestValidator_AcceptsHTTP
--- PASS: TestValidator_AcceptsHTTP (0.00s)
=== RUN TestValidator_RejectsFTP
--- PASS: TestValidator_RejectsFTP (0.00s)
=== RUN TestValidator_RejectsNoScheme
--- PASS: TestValidator_RejectsNoScheme (0.00s)
=== RUN TestValidator_RejectsTooLong
--- PASS: TestValidator_RejectsTooLong (0.00s)
=== RUN TestValidator_RejectsEmpty
--- PASS: TestValidator_RejectsEmpty (0.00s)
=== RUN TestValidator_RejectsSliceOverLimit
--- PASS: TestValidator_RejectsSliceOverLimit (0.00s)
=== RUN TestStore_DuplicateSKU
--- PASS: TestStore_DuplicateSKU (0.00s)
=== RUN TestStore_ListDoesNotLoadMediaArrays
--- PASS: TestStore_ListDoesNotLoadMediaArrays (0.00s)
=== RUN TestStore_PaginationOffset
--- PASS: TestStore_PaginationOffset (0.00s)
=== RUN TestStore_OffsetBeyondTotal
--- PASS: TestStore_OffsetBeyondTotal (0.00s)
=== RUN TestHTTP_CreateProduct_201
--- PASS: TestHTTP_CreateProduct_201 (0.00s)
=== RUN TestHTTP_CreateProduct_EmptyName
--- PASS: TestHTTP_CreateProduct_EmptyName (0.00s)
=== RUN TestHTTP_CreateProduct_EmptySKU
--- PASS: TestHTTP_CreateProduct_EmptySKU (0.00s)
=== RUN TestHTTP_CreateProduct_DuplicateSKU_409
--- PASS: TestHTTP_CreateProduct_DuplicateSKU_409 (0.00s)
--- PASS: TestHTTP_CreateProduct_InvalidURL (0.00s)
=== RUN TestHTTP_ListProducts_DefaultPagination
--- PASS: TestHTTP_ListProducts_DefaultPagination (0.00s)
=== RUN TestHTTP_ListProducts_Pagination
--- PASS: TestHTTP_ListProducts_Pagination (0.00s)
=== RUN TestHTTP_GetProduct_FullMedia
--- PASS: TestHTTP_GetProduct_FullMedia (0.00s)
=== RUN TestHTTP_GetProduct_NotFound
--- PASS: TestHTTP_GetProduct_NotFound (0.00s)
=== RUN TestHTTP_AddMedia_AppendsURLs
--- PASS: TestHTTP_AddMedia_AppendsURLs (0.00s)
=== RUN TestHTTP_AddMedia_EmptyBody_400
--- PASS: TestHTTP_AddMedia_EmptyBody_400 (0.00s)
=== RUN TestHTTP_AddMedia_NotFound
--- PASS: TestHTTP_AddMedia_NotFound (0.00s)
=== RUN TestPerformanceInvariant_ListDoesNotSerialiseAllMedia
--- PASS: TestPerformanceInvariant_ListDoesNotSerialiseAllMedia (0.00s)
=== RUN TestConcurrentCreate
--- PASS: TestConcurrentCreate (0.00s)
PASS
ok source-asia-backend-assignment/part2 0.413s
? source-asia-backend-assignment/part2/handlers [no test files]
? source-asia-backend-assignment/part2/models [no test files]
? source-asia-backend-assignment/part2/store [no test files]
? source-asia-backend-assignment/part2/validator [no test files]
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> go test -race
=== RUN TestValidator_AcceptsHTTPS
--- PASS: TestValidator_AcceptsHTTPS (0.00s)
=== RUN TestValidator_AcceptsHTTP
--- PASS: TestValidator_AcceptsHTTP (0.00s)
=== RUN TestValidator_RejectsFTP
--- PASS: TestValidator_RejectsFTP (0.00s)
=== RUN TestValidator_RejectsNoScheme
--- PASS: TestValidator_RejectsNoScheme (0.00s)
=== RUN TestValidator_RejectsTooLong
--- PASS: TestValidator_RejectsTooLong (0.00s)
=== RUN TestValidator_RejectsEmpty
--- PASS: TestValidator_RejectsEmpty (0.00s)
=== RUN TestValidator_RejectsSliceOverLimit
--- PASS: TestValidator_RejectsSliceOverLimit (0.00s)
=== RUN TestStore_DuplicateSKU
--- PASS: TestStore_DuplicateSKU (0.00s)
=== RUN TestStore_ListDoesNotLoadMediaArrays
--- PASS: TestStore_ListDoesNotLoadMediaArrays (0.00s)
=== RUN TestStore_PaginationOffset
--- PASS: TestStore_PaginationOffset (0.00s)
=== RUN TestStore_OffsetBeyondTotal
--- PASS: TestStore_OffsetBeyondTotal (0.00s)
=== RUN TestHTTP_CreateProduct_201
--- PASS: TestHTTP_CreateProduct_201 (0.00s)
=== RUN TestHTTP_CreateProduct_EmptyName
--- PASS: TestHTTP_CreateProduct_EmptyName (0.00s)
--- PASS: TestHTTP_CreateProduct_EmptySKU (0.00s)
=== RUN TestHTTP_CreateProduct_DuplicateSKU_409
--- PASS: TestHTTP_CreateProduct_DuplicateSKU_409 (0.00s)
=== RUN TestHTTP_CreateProduct_InvalidURL
=== RUN TestHTTP_ListProducts_DefaultPagination
--- PASS: TestHTTP_ListProducts_DefaultPagination (0.01s)
=== RUN TestHTTP_ListProducts_Pagination
--- PASS: TestHTTP_ListProducts_Pagination (0.01s)
=== RUN TestHTTP_GetProduct_FullMedia
--- PASS: TestHTTP_GetProduct_FullMedia (0.00s)
=== RUN TestHTTP_GetProduct_NotFound
--- PASS: TestHTTP_GetProduct_NotFound (0.00s)
=== RUN TestHTTP_AddMedia_AppendsURLs
--- PASS: TestHTTP_AddMedia_AppendsURLs (0.00s)
=== RUN TestHTTP_AddMedia_EmptyBody_400
--- PASS: TestHTTP_AddMedia_EmptyBody_400 (0.00s)
=== RUN TestHTTP_AddMedia_NotFound
--- PASS: TestHTTP_AddMedia_NotFound (0.00s)
=== RUN TestPerformanceInvariant_ListDoesNotSerialiseAllMedia
--- PASS: TestPerformanceInvariant_ListDoesNotSerialiseAllMedia (0.01s)
=== RUN TestConcurrentCreate
--- PASS: TestConcurrentCreate (0.00s)
PASS
ok source-asia-backend-assignment/part2 1.775s
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> go test ./part
=== RUN TestPerformanceInvariant_ListDoesNotSerialiseAllMedia
--- PASS: TestPerformanceInvariant_ListDoesNotSerialiseAllMedia (0.00s)
PASS
ok source-asia-backend-assignment/part2 0.421s
PS C:\Users\Shahaan\Documents\Sourceasia_Assignment\source-asia-backend-assignment> go test ./part
=== RUN TestConcurrentCreate
--- PASS: TestConcurrentCreate (0.00s)
PASS
ok source-asia-backend-assignment/part2 0.467s
