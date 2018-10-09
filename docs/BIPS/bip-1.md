---
bip: 1
title: BIP Purpose and Guidelines
discussions-to: https://chat.hyperledger.org/channel/burrow-contributors
status: Draft
type: Meta
author: The Burrow's marmots and others
        https://github.com/hyperledger/burrow/BIPs/blob/master/BIPS/bip-1.md
created: 2018-10-09
---

## What is a BIP?

BIP stands for Burrow Improvement Proposal. A BIP is a design document providing information to the Burrow community, or describing a new feature for Burrow or its processes or environment. The BIP should provide a concise technical specification of the feature and a rationale for the feature. The BIP author is responsible for building consensus within the community and documenting dissenting opinions.

![burrow logo](../assets/bip-1/burrow-logo.png)

## BIP Rationale

We intend BIPs to be the primary mechanisms for proposing new features, for collecting community technical input on an issue, and for documenting the design decisions that have gone into Burrow. Because the BIPs are maintained as text files in a versioned repository, their revision history is the historical record of the feature proposal.

## BIP Formats and Templates

BIPs should be written in [markdown] format.
Image files should be included in a subdirectory of the `assets` folder for that BIP as follow: `assets/bip-X` (for bip **X**). When linking to an image in the BIP, use relative links such as `../assets/bip-X/image.png`.

## BIP Header Preamble

Each BIP must begin with an RFC 822 style header preamble, preceded and followed by three hyphens (`---`). The headers must appear in the following order. Headers marked with "*" are optional and are described below. All other headers are required.

` bip:` <BIP number> (this is determined by the BIP editor)

` title:` <BIP title>

` author:` <a list of the author's or authors' name(s) and/or username(s), or name(s) and email(s). Details are below.>

` * discussions-to:` \<a url pointing to the official discussion thread\>

 - :x: `discussions-to` can not be a Github `Pull-Request`.

` status:` <Draft | Last Call | Accepted | Final | Active | Deferred | Rejected | Superseded>

`* review-period-end:` YYYY-MM-DD

` type:` <Standards Track (Accounts and State, Consensus, Crypto, EVM, Gateway)  | Informational | Meta>

` * category:` <Accounts and State | Consensus | Crypto | EVM | Gateway>

` created:` <date created on, in ISO 8601 (yyyy-mm-dd) format>

` * requires:` <BIP number(s)>

` * replaces:` <BIP number(s)>

` * superseded-by:` <BIP number(s)>

` * resolution:` \<a url pointing to the resolution of this BIP\>

#### Author header

The author header optionally lists the names, email addresses or usernames of the authors/owners of the BIP. Those who prefer anonymity may use a username only, or a first name and a username. The format of the author header value must be:

Random J. User &lt;address@dom.ain&gt;

or

Random J. User (@username)

if the email address or GitHub username is included, and

Random J. User

if the email address is not given.

Note: The resolution header is required for Standards Track BIPs only. It contains a URL that should point to an email message or other web resource where the pronouncement about the BIP is made.

While a BIP is a draft, a discussions-to header will indicate the mailing list or URL where the BIP is being discussed.

The type header specifies the type of BIP: Standards Track, Meta, or Informational. If the track is Standards please include the subcategory (core, networking, consensus, or EVM).

The category header specifies the BIP's category. This is required for standards-track BIPs only.

The created header records the date that the BIP was assigned a number. Both headers should be in yyyy-mm-dd format, e.g. 2001-08-14.

BIPs may have a requires header, indicating the BIP numbers that this BIP depends on.

BIPs may also have a superseded-by header indicating that a BIP has been rendered obsolete by a later document; the value is the number of the BIP that replaces the current document. The newer BIP must have a Replaces header containing the number of the BIP that it rendered obsolete.

Headers that permit lists must separate elements with commas.

## History

This document was derived heavily from [Ethereum's EIP1] written by Martin Becze and Hudson Jameson which in turn was derived from [Bitcoin's BIP-0001] written by Amir Taaki which in turn was derived from [Python's PEP-0001]. In many places text was simply copied and modified. Although the PEP-0001 text was written by Barry Warsaw, Jeremy Hylton, and David Goodger, they are not responsible for its use in the Burrow Improvement Process, and should not be bothered with technical questions specific to Burrow or the BIP. Please direct all comments to the BIP editors.
