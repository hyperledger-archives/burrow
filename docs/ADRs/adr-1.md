---
adr: 1
title: ADR Purpose and Guidelines
discussions-to: https://chat.hyperledger.org/channel/burrow-contributors
status: Draft
type: Meta
author: The Burrow's marmots and others
        https://github.com/hyperledger/burrow/ADRs/blob/main/ADRs/adr-1.md
created: 2018-10-09
---

## What is an ADR?

ADR stands for Architecture Decision Record.
An ADR is a design document providing information to the Burrow community, or describing a new feature for Burrow or its processes or environment.
The ADR should provide a concise technical specification of the feature and a rationale for the feature.
The ADR author is responsible for building consensus within the community and documenting dissenting opinions.

![burrow logo](../assets/adr-1/burrow-logo.png)

## ADR Rationale

We intend ADRs to be the primary mechanisms for proposing new features, for collecting community technical input on an issue, and for documenting the design decisions that have gone into Burrow. Because the ADRs are maintained as text files in a versioned repository, their revision history is the historical record of the feature proposal.

## ADR Formats and Templates

ADRs should be written in [markdown] format.
Image files should be included in a subdirectory of the `assets` folder for that ADR as follow: `assets/adr-X` (for adr **X**). When linking to an image in the ADR, use relative links such as `../assets/adr-X/image.png`.

## ADR Header Preamble

Each ADR must begin with an RFC 822 style header preamble, preceded and followed by three hyphens (`---`). The headers must appear in the following order. Headers marked with "*" are optional and are described below. All other headers are required.

` adr:` <ADR number>

` title:` <ADR title>

` author:` <a list of the author's or authors' name(s) and/or username(s), or name(s) and email(s). Details are below.>

` * discussions-to:` \<a url pointing to the official discussion thread\>

 - :x: `discussions-to` can not be a Github `Pull-Request`.

` status:` <Draft | Last Call | Accepted | Final | Active | Deferred | Rejected | Superseded>

`* review-period-end:` YYYY-MM-DD

` type:` <Standards Track (Accounts and State, Consensus, Crypto, EVM, Gateway)  | Informational | Meta>

` * category:` <Accounts and State | Consensus | Crypto | EVM | Gateway>

` created:` <date created on, in ISO 8601 (yyyy-mm-dd) format>

` * requires:` <ADR number(s)>

` * replaces:` <ADR number(s)>

` * superseded-by:` <ADR number(s)>

` * resolution:` \<a url pointing to the resolution of this ADR\>

#### Author header

The author header optionally lists the names, email addresses or usernames of the authors/owners of the ADR. Those who prefer anonymity may use a username only, or a first name and a username. The format of the author header value must be:

Random J. User &lt;address@dom.ain&gt;

or

Random J. User (@username)

if the email address or GitHub username is included, and

Random J. User

if the email address is not given.

Note: The resolution header is required for Standards Track ADRs only. It contains a URL that should point to an email message or other web resource where the pronouncement about the ADR is made.

While an ADR is a draft, a discussions-to header will indicate the mailing list or URL where the ADR is being discussed.

The type header specifies the type of ADR: Standards Track, Meta, or Informational. If the track is Standards please include the subcategory (core, networking, consensus, or EVM).

The category header specifies the ADR's category. This is required for standards-track ADRs only.

The created header records the date that the ADR was assigned a number. Both headers should be in yyyy-mm-dd format, e.g. 2001-08-14.

ADRs may have a requires header, indicating the ADR numbers that this ADR depends on.

ADRs may also have a superseded-by header indicating that an ADR has been rendered obsolete by a later document; the value is the number of the ADR that replaces the current document. The newer ADR must have a Replaces header containing the number of the ADR that it rendered obsolete.

Headers that permit lists must separate elements with commas.

## History

This document was derived heavily from [Ethereum's EIP1] written by Martin Becze and Hudson Jameson which in turn was derived from [Bitcoin's ADR-0001] written by Amir Taaki which in turn was derived from [Python's PEP-0001]. In many places text was simply copied and modified. Although the PEP-0001 text was written by Barry Warsaw, Jeremy Hylton, and David Goodger, they are not responsible for its use in the Burrow Improvement Process, and should not be bothered with technical questions specific to Burrow or the ADR. Please direct all comments to the ADR editors.

This document was inspired from [ADR GitHub Organization](https://adr.github.io/) and [ADR Examples](https://github.com/joelparkerhenderson/architecture_decision_record).
