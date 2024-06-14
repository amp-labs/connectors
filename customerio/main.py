if __name__ == '__main__':
    import json
    # Opening JSON file
    f = open('cio_journeys_app_api.json')

    # returns JSON object as
    # a dictionary
    result = {}

    data = json.load(f)
    paths = data["paths"]
    for k, v in paths.items():
        for method_name, v2 in v.items():
            if "description" in v2:
                if method_name in result:
                    result[method_name].append(k)
                else:
                    result[method_name] = [k]
                print(k, method_name)

    for k, v in result.items():
        print()
        print(k)
        for i in range(len(v)):
            print(v[i])

