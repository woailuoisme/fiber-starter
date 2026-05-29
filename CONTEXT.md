# Fiber Starter

This context defines the domain language for reusable application capabilities in this backend starter.

## Language

**Media Library**:
A reusable capability for associating uploaded files with application records. In this project, the Media Library starts as a simplified core focused on stored media, collections, and derived media.
_Avoid_: file uploader, attachment helper

**Media**:
A stored file that belongs to one application record and may have derived versions. Media includes its collection, file identity, storage location, and descriptive properties.
_Avoid_: file, attachment, asset

**Media Disk**:
The storage location used by the Media Library. The simplified core uses one default Media Disk while each Media item still records where it was stored.
_Avoid_: filesystem choice, storage backend option

**Media Path**:
The stable storage location for Media and its Derived Media. The simplified core stores original Media under a UUID-based path and keeps Derived Media under that Media item's conversions path.
_Avoid_: file path string, directory layout

**Media Owner**:
An application record that owns Media through a polymorphic identity. One Media Owner may have many Media items across multiple Media Collections.
_Avoid_: parent model, related entity

**Media Collection**:
A named group of Media for one application record, such as an avatar or gallery. In the simplified core, a collection may define single-file ownership, accepted MIME types, maximum file size, and Derived Media rules.
_Avoid_: folder, category, bucket

**Derived Media**:
A generated version of a Media item, such as a thumbnail or converted image. Derived Media belongs to its original Media and is not a separate Media record or user upload.
_Avoid_: conversion file, thumbnail file

**Derived Media Status**:
The availability state of Derived Media for a Media item. In the simplified core, Derived Media may be pending, completed, or failed.
_Avoid_: conversion flag, generated map

## Example Dialogue

Dev: "Should this product image use the Media Library?"

Domain expert: "Yes. The product is the Media Owner. Store the original upload as Media in the product's gallery Media Collection, then generate Derived Media for previews."

Dev: "Is the preview a new Media item?"

Domain expert: "No. It is Derived Media for the original Media."

Dev: "Can the UI use the preview immediately?"

Domain expert: "Only when the Derived Media Status is completed. If it is pending, use the original Media or wait for generation to finish."
