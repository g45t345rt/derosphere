from ast import operator
import csv
import json
import sys
import operator

metadata_path = "./metadata.json"  # sys.argv[1]
rarity_path = "./rarity.csv"  # sys.argv[2]
nfts_out_path = "./nfts.json"  # sys.argv[3]
collection_stats_out_path = "./stats.json"  # sys.argv[3]
ipfs_folder_cid = "QmP3HnzWpiaBA6ZE8c3dy5ExeG7hnYjSqkNfVbeVW5iEp6"


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


def captain_nft(id, index):
    return {
        "id": id,
        "name": "Captain #{i}".format(i=index),
        "image": "ipfs://{cid}/low/captain.jpg".format(cid=ipfs_folder_cid),
        "attributes": {},
        "score": 100
    }


def jeff_nft(id):
    return {
        "id": id,
        "name": "Jeff",
        "image": "ipfs://{cid}/low/jeff.jpg".format(cid=ipfs_folder_cid),
        "attributes": {},
        "score": 100
    }


def main():
    metadata_file = open(metadata_path, "rb")
    metadata = json.load(metadata_file)
    metadata_file.close()

    attributes_names_file = open("./attributes_names.json")
    attributes_names = json.load(attributes_names_file)
    attributes_names_file.close()

    # rarity = get_rarity() // don't need rarity file anymore - the nft platform will calculate it
    collection_stats = {}
    #attr_stats = {}
    nfts = []
    for cI in range(len(metadata["collection"])):
        collection = metadata["collection"][cI]
        id = collection["name"].replace("#", "")  # don't need # with ID

        nft = {}
        nft["id"] = int(id)
        nft["name"] = "Dero Seals #{id}".format(id=id)
        nft["image"] = "ipfs://{cid}/low/{id}.jpg".format(
            cid=ipfs_folder_cid, id=id)
        #nft["rarity"] = float(rarity[id])
        #nft["rarity_type"] = get_rarity_type(nft["rarity"])
        nft["attributes"] = {}

        attributes = collection["attributes"]
        is_dero_man_suit = False
        for attribute in attributes:
            attr_category = capitalize_words(attribute["trait_type"])
            if attr_category == "Background" or attr_category == "Base":
                continue
            attr_name = attributes_names[attribute["value"].replace(
                "Untitled_Artwork ", "")]  # Remove unnecessary Untitled_Artwork
            if attr_category == "Shirts":
                is_dero_man_suit = attr_name == "Dero Man Suit"
            nft["attributes"][attr_category] = attr_name
            if attr_category not in collection_stats:
                collection_stats[attr_category] = {
                    "count": 0, "attributes": {}}
            collection_stats[attr_category]["count"] += 1
            if attr_name not in collection_stats[attr_category]["attributes"]:
                collection_stats[attr_category]["attributes"][attr_name] = {
                    "count": 0}
            collection_stats[attr_category]["attributes"][attr_name]["count"] += 1
            #attr_stats[attr_name] = attr_stats.get(attr_name, 0) + 1

        # Metadata fix - All Dero Man Suit should have Blue Eyes
        if is_dero_man_suit:
            if "Eyes" in nft["attributes"]:
                if nft["attributes"]["Eyes"] in ["Green Eyes", "Red Eyes", "Blue Eyes"]:
                    nft["attributes"]["Eyes"] = "Blue Eyes"
            else:
                nft["attributes"]["Eyes"] = "Blue Eyes"
        nfts.append(nft)

    nft_count = len(nfts)
    for layer in collection_stats:
        c_count = collection_stats[layer]["count"]
        collection_stats[layer]["percentage"] = round(
            c_count * 100 / nft_count, 2)
        for attribute in collection_stats[layer]["attributes"]:
            count = collection_stats[layer]["attributes"][attribute]["count"]
            collection_stats[layer]["attributes"][attribute]["percentage"] = round(
                count * 100 / nft_count, 2)
            collection_stats[layer]["attributes"][attribute]["score"] = round(
                1 / (count / nft_count), 2)

        n_count = nft_count - c_count
        collection_stats[layer]["attributes"]["None"] = {
            "count": n_count, "percentage": round(n_count * 100 / nft_count, 2), "score": 0}
        # for nft in nfts:
        # if layer not in nft["attributes"]:
        #  nft["attributes"][layer] = "None"

    for nft in nfts:
        nft_score = 0
        for layer in nft["attributes"]:
            attr = nft["attributes"][layer]
            a_score = collection_stats[layer]["attributes"][attr]["score"]
            nft_score += a_score
        nft["score"] = round(nft_score, 2)

    nfts.sort(key=operator.itemgetter("score"), reverse=True)

    # remove last 10 nfts with 9 Captain and Jeff
    for i in range(1, 10):
        pos = len(nfts)-11+i
        nft = nfts[pos]
        nfts[pos] = captain_nft(nft["id"], i)
    nfts[3499] = jeff_nft(nfts[3499]["id"])

    nfts_file = open(nfts_out_path, "w")
    json.dump(nfts, nfts_file, indent=2)
    nfts_file.close()

    stats_file = open(collection_stats_out_path, "w")
    json.dump(collection_stats, stats_file, indent=2)
    stats_file.close()


if __name__ == "__main__":
    main()
