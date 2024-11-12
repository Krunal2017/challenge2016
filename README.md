# Real Image Challenge 2016

In the cinema business, a feature film is usually provided to a regional distributor based on a contract for exhibition in a particular geographical territory.

Each authorization is specified by a combination of included and excluded regions. For example, a distributor might be authorzied in the following manner:
```
Permissions for DISTRIBUTOR1
INCLUDE: INDIA
INCLUDE: UNITEDSTATES
EXCLUDE: KARNATAKA-INDIA
EXCLUDE: CHENNAI-TAMILNADU-INDIA
```
This allows `DISTRIBUTOR1` to distribute in any city inside the United States and India, *except* cities in the state of Karnataka (in India) and the city of Chennai (in Tamil Nadu, India).

At this point, asking your program if `DISTRIBUTOR1` has permission to distribute in `CHICAGO-ILLINOIS-UNITEDSTATES` should get `YES` as the answer, and asking if distribution can happen in `CHENNAI-TAMILNADU-INDIA` should of course be `NO`. Asking if distribution is possible in `BANGALORE-KARNATAKA-INDIA` should also be `NO`, because the whole state of Karnataka has been excluded.

Sometimes, a distributor might split the work of distribution amount smaller sub-distiributors inside their authorized geographies. For instance, `DISTRIBUTOR1` might assign the following permissions to `DISTRIBUTOR2`:

```
Permissions for DISTRIBUTOR2 < DISTRIBUTOR1
INCLUDE: INDIA
EXCLUDE: TAMILNADU-INDIA
```
Now, `DISTRIBUTOR2` can distribute the movie anywhere in `INDIA`, except inside `TAMILNADU-INDIA` and `KARNATAKA-INDIA` - `DISTRIBUTOR2`'s permissions are always a subset of `DISTRIBUTOR1`'s permissions. It's impossible/invalid for `DISTRIBUTOR2` to have `INCLUDE: CHINA`, for example, because `DISTRIBUTOR1` isn't authorized to do that in the first place. 

If `DISTRIBUTOR2` authorizes `DISTRIBUTOR3` to handle just the city of Hubli, Karnataka, India, for example:
```
Permissions for DISTRIBUTOR3 < DISTRIBUTOR2 < DISTRIBUTOR1
INCLUDE: HUBLI-KARNATAKA-INDIA
```
Again, `DISTRIBUTOR2` cannot authorize `DISTRIBUTOR3` with a region that they themselves do not have access to. 

We've provided a CSV with the list of all countries, states and cities in the world that we know of - please use the data mentioned there for this program. *The codes you see there may be different from what you see here, so please always use the codes in the CSV*. This Readme is only an example. 

Write a program in any language you want (If you're here from Gophercon, use Go :D) that does this. Feel free to make your own input and output format / command line tool / GUI / Webservice / whatever you want. Feel free to hold the dataset in whatever structure you want, but try not to use external databases - as far as possible stick to your langauage without bringing in MySQL/Postgres/MongoDB/Redis/Etc.

To submit a solution, fork this repo and send a Pull Request on Github. 

For any questions or clarifications, raise an issue on this repo and we'll answer your questions as fast as we can.

## Solution Description
The service starts an HTTP server to listen to incoming requests for creating distributors. It expects to receive the distributor data from POST request in the JSON format as given below:
```
{
  "name": "DISTRIBUTOR2",
        "inherits": "DISTRIBUTOR1",
        "include": ["IN"],
        "exclude": ["JK:IN"]
}
```
The location codes follow the format `<city>:<province>:<country>`. To include the entire province or country, codes such as `<province>:<country>` or `<country>` could be used.

The service can handle GET and POST requests currently, to perform the following operations:
1. GET requests: Can be used to query whether a distributor has access to a certain location code.
2. POST requests: Can be used to create new distributors. It returns a JSON response containing the data it has stored. It weeds out any locations the distributor cannot have access to.

In future following requests could also be implemented:
1. PATCH request: Can be used to update the include or exclude lists of a distributor
2. PUT request: Can be used to replace the existing distributor

### Limitations:
Currently, a distributor cannot be updated once it is created via POST request. 

### GET request example
```
curl "http://localhost:8080/distributor?distributor=DISTRIBUTOR1&location=UDHAP:JK:IN"
```

### POST request example
```
curl -X POST http://localhost:8080/distributor -H "Content-Type: application/json" -d '{
  "name": "DISTRIBUTOR1",
    "include": ["IN", "WS"],
    "exclude": ["UP:IN", "YAVTM:MH:IN"]
}'
```

### Expected responses
StatusOK - 200 : Distributor can access the given location
StatusCreated - 201 : Successfully created Distributor
StatusNotFound - 404 : Distributor not found
StatusForbidden - 403 :  Distributor cannot access the given location
StatusBadRequest - 400 : Received invalid JSON input
StatusConflict - 409 : Distributor already exists
StatusInternalServerError - 500 : Error occurred while returning response

## Run Unit tests

Unit tests can be run using the command:
go test

## Bulk/Stress test:

The script test.py can be used as follows:
python3 test.py

This script can be used to generate a chain of 50 distributors by default. Script can be modified to test more number of distributors or increasing the number of cities being included by the distributors.