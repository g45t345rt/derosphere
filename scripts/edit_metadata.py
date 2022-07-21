import csv
import json
import sys

metadata_path = sys.argv[1]
rarity_path = sys.argv[2]


def get_rarity():
    rarity_file = open(rarity_path, newline="")
    csv_rarity = csv.reader(rarity_file, delimiter=",")
    rarity = {}
    for row in csv_rarity:
        fileNumber = row[0]
        rarityValue = row[1]
        rarity[fileNumber] = rarityValue
    return rarity


def main():
    metadata_file = open(metadata_path, "rb")
    data = json.load(metadata_file)
    rarity = get_rarity()
    for c in data["collection"]:
        fileNumber = c["name"].replace("#", "")
        c["attributes"].append(
            {"trait_type": "id", "value": fileNumber})
        c["attributes"].append(
            {"trait_type": "rarity", "value": rarity[fileNumber]})

    metadata_file.close
    metadata_file = open(metadata_path, "w")
    json.dump(data, metadata_file, indent=2)
    metadata_file.close()


if __name__ == "__main__":
    main()
