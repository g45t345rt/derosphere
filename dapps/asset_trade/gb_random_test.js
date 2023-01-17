// random value from disposable asset with cursor increment if asset is already disposed
// can't use array in dero dvm so this is my way of working around it and make the gumball random machine works with a lot of assets inside

const assets = []
const max = 10
for (let i = 0; i < max; i++) {
  assets.push(0)
}

let cursor = 0
let reads = []
let results = []
for (let i = 0; i < max; i++) {
  const rand = Math.floor(Math.random() * assets.length)
  if (assets[rand] === 1) {
    for (let a=cursor;a<max;a++) {
      if (assets[a] === 0) {
        assets[a] = 1
        reads.push(a-cursor+1)
        results.push(a)
        cursor = a
        break
      }
    }
  } else {
    reads.push(1)
    results.push(rand)
    assets[rand] = 1
  }
}

console.log(`assets`)
console.log(assets)

console.log(`read count`)
console.log(reads) //.sort((a,b) => b - a))

console.log(`results`)
console.log(results)

