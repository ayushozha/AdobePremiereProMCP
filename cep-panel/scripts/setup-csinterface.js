#!/usr/bin/env node

/**
 * Downloads CSInterface.js from Adobe's CEP-Resources GitHub repository
 * and places it in src/CSInterface.js, replacing the placeholder.
 */

var https = require("https");
var fs = require("fs");
var path = require("path");

var CSINTERFACE_URL =
    "https://raw.githubusercontent.com/Adobe-CEP/CEP-Resources/master/CEP_11.x/CSInterface.js";

var OUTPUT_PATH = path.join(__dirname, "..", "src", "CSInterface.js");

console.log("Downloading CSInterface.js from Adobe CEP-Resources...");
console.log("  URL: " + CSINTERFACE_URL);
console.log("  Destination: " + OUTPUT_PATH);

function download(url, dest) {
    return new Promise(function (resolve, reject) {
        var file = fs.createWriteStream(dest);
        https.get(url, function (response) {
            // Handle redirects
            if (response.statusCode === 301 || response.statusCode === 302) {
                file.close();
                fs.unlinkSync(dest);
                download(response.headers.location, dest).then(resolve).catch(reject);
                return;
            }

            if (response.statusCode !== 200) {
                file.close();
                fs.unlinkSync(dest);
                reject(new Error("HTTP " + response.statusCode + " from " + url));
                return;
            }

            response.pipe(file);
            file.on("finish", function () {
                file.close();
                resolve();
            });
        }).on("error", function (err) {
            file.close();
            try { fs.unlinkSync(dest); } catch (e) { /* ignore */ }
            reject(err);
        });
    });
}

download(CSINTERFACE_URL, OUTPUT_PATH)
    .then(function () {
        var stat = fs.statSync(OUTPUT_PATH);
        console.log("  Downloaded successfully (" + stat.size + " bytes)");
    })
    .catch(function (err) {
        console.error("  Failed to download CSInterface.js: " + err.message);
        console.error("");
        console.error("  You can manually download it from:");
        console.error("    https://github.com/Adobe-CEP/CEP-Resources/tree/master/CEP_11.x");
        console.error("  And place it at:");
        console.error("    " + OUTPUT_PATH);
        process.exit(1);
    });
