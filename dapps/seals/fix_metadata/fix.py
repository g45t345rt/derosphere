import csv
import json
import sys

metadata_path = "./metadata.json"  # sys.argv[1]
rarity_path = "./rarity.csv"  # sys.argv[2]
out_path = "./nfts.json"  # sys.argv[3]


def get_rarity():
    rarity_file = open(rarity_path, newline="")
    csv_rarity = csv.reader(rarity_file, delimiter=",")
    rarity = {}
    for row in csv_rarity:
        fileNumber = row[0]
        rarityValue = row[1]
        rarity[fileNumber] = rarityValue
    return rarity


def get_rarity_type(rarity):
    if (rarity >= 353.62):
        return "ultra_rare"
    elif (rarity >= 300.11):
        return "rare"
    else:
        return "common"


def capitalize_words(value):
    words = []
    for v in value.split(" "):
        words.append(v.capitalize())
    return " ".join(words)


def main():
    metadata_file = open(metadata_path, "rb")
    metadata = json.load(metadata_file)
    metadata_file.close()

    attributes_names_file = open("./attributes_names.json")
    attributes_names = json.load(attributes_names_file)
    attributes_names_file.close()

    rarity = get_rarity()
    nfts = []
    for cI in range(len(metadata["collection"])):
        collection = metadata["collection"][cI]
        id = collection["name"].replace("#", "")

        nfts.append({})
        nft = nfts[cI]
        nft["id"] = int(id)
        nft["rarity"] = float(rarity[id])
        nft["rarity_type"] = get_rarity_type(nft["rarity"])
        nft["attributes"] = {}

        attributes = collection["attributes"]
        for attribute in attributes:
            attr_category = capitalize_words(attribute["trait_type"])
            if attr_category == "Background" or attr_category == "Base":
                continue
            attr_name = attributes_names[attribute["value"].replace(
                "Untitled_Artwork ", "")]
            nft["attributes"][attr_category] = attr_name

    nfts_file = open(out_path, "w")
    json.dump(nfts, nfts_file, indent=2)
    nfts_file.close()


if __name__ == "__main__":
    main()
