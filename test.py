import csv
import requests
import random
import json

BASE_URL = 'http://localhost:8080'  # server's URL

country_db = {}
distributor_db = {}

# Generate list of locations for distributors with no parents
def generate_random_list():
    # Generate upto 100 random locations
    list_length = random.randint(1, 100)
    locations_set = set()

    while len(locations_set) < list_length:
        country_code = random.choice(list(country_db.keys()))
        province_code = random.choice(list(country_db[country_code].keys()))
        city_code = random.choice(list(country_db[country_code][province_code].keys()))
        choice = random.choice([1, 2, 3])

        if choice == 1:
            location = f"{city_code}:{province_code}:{country_code}"
        elif choice == 2:
            location = f"{province_code}:{country_code}"
        else:
            location = country_code

        locations_set.add(location)

    return list(locations_set)

# Generate list from include list of distributor or it's parent
def generate_from_list(include_list):
    matched_locations = []
    k = random.randint(1, len(include_list))
    sub_list = random.sample(include_list, k)

    for location in sub_list:
        parts = location.split(":")

        if len(parts) == 1:
            country = parts[0]
            if country in country_db:
                province = random.choice(list(country_db[country].keys()))
                city = random.choice(list(country_db[country][province].keys()))
                matched_locations.append(f"{city}:{province}:{country}")

        elif len(parts) == 2:
            province, country = parts
            if country in country_db and province in country_db[country]:
                city = random.choice(list(country_db[country][province].keys()))
                matched_locations.append(f"{city}:{province}:{country}")

        elif len(parts) == 3:
            matched_locations.append(location)

    return matched_locations

# Test GET and POST requests by creating distributors in bulk
def bulk_test_get_and_post():
    NUM_DISTRIBUTORS = 50
    for i in range(1, NUM_DISTRIBUTORS+1):
        distributor_name = f"distributor-{i}"

        # Determine the range for selecting parent distributors
        if i <= 10:
            include_list = generate_random_list()
        else:
            # Calculate the previous range
            min_parent = ((i - 1) // 10 - 1) * 10 + 1
            max_parent = min_parent + 9
            # Select a random parent from previous range
            parent_num = random.randint(min_parent, max_parent)
            parent_name = f"distributor-{parent_num}"

            # Ensure the parent distributor exists
            if parent_name in distributor_db:
                parent_list = distributor_db[parent_name].get("Include", [])
                include_list = generate_from_list(parent_list)
            else:
                include_list = generate_random_list()

        exclude_list = generate_from_list(include_list)
        exclude_list = [code for code in exclude_list if code not in include_list]

        print(f"Distributor Name: {distributor_name}")
        print("Include List:", include_list)
        print("Exclude List:", exclude_list)

        response_text = test_post_request(distributor_name, include_list, exclude_list)
        response_dict = json.loads(response_text)
        distributor_db[distributor_name] = response_dict

        include_list = response_dict.get("Include", [])
        exclude_list = response_dict.get("Exclude", [])
        filtered_include_list = [code for code in include_list if code not in exclude_list]
        
        if filtered_include_list:
            location_code = random.choice(filtered_include_list)
            response = test_get_request(distributor_name, location_code)
            assert response.status_code == 200
            assert response.text == "YES"

        if exclude_list:
            location_code = random.choice(exclude_list)
            response = test_get_request(distributor_name, location_code)
            assert response.status_code == 403
            assert response.text == "NO"

def test_post_request(distributor_name, include_list, exclude_list):
    print(f"POST Request: {distributor_name}, Include={include_list}, Exclude={exclude_list}")
    request_data = {
        "name": distributor_name,
        "include": include_list,
        "exclude": exclude_list,
        "inherits": ""
    }
    response = requests.post(f"{BASE_URL}/distributor", json=request_data)
    return response.text

def test_get_request(distributor, location_code):
    response = requests.get(f"{BASE_URL}/distributor?distributor={distributor}&location={location_code}")
    print(f"GET Request: {distributor}, Location={location_code}, Response:{response.text}")
    return response

def parse_db(data):
    for row in data:
        if row:
            city_code, city = row[0], row[3]
            province_code = row[1]
            country_code = row[2]

            if country_code not in country_db:
                country_db[country_code] = {}

            if province_code not in country_db[country_code]:
                country_db[country_code][province_code] = {}

            country_db[country_code][province_code][city_code] = city

def main():
    csv_file_path = 'cities.csv'
    with open(csv_file_path, newline='', encoding='utf-8') as csvfile:
        csvreader = csv.reader(csvfile)
        header = next(csvreader)
        data = [row for row in csvreader if row]

    parse_db(data)
    bulk_test_get_and_post()

if __name__ == "__main__":
    main()
