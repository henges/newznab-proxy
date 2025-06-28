package newznab_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/henges/newznab-proxy/newznab"
	"github.com/henges/newznab-proxy/xmlutil"
	"github.com/stretchr/testify/assert"
)

const testXml = `
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:newznab="http://www.newznab.com/DTD/2010/feeds/attributes/">
  <channel>
    <atom:link href="https://api.test.com/api?t=search&amp;q=test%20test&amp;sort=size_asc" rel="self" type="application/rss+xml" />
    <title>test.com</title>
    <description>API Feed</description>
    <link>https://api.test.com/</link>
    <language>en-gb</language>
    <webMaster>root@test.com (test.com)</webMaster>
    <category />
    <image>
      <url>https://api.test.com/templates/default/images/banner.jpg</url>
      <title>test.com</title>
      <link>https://api.test.com/</link>
      <description>Visit test.com - </description>
    </image>
    <newznab:response offset="0" total="1" />
    <item>
      <title>Test Test</title>
      <guid isPermaLink="true">https://api.test.com/details/1efe314025c6661380c7edf9938c38b3</guid>
      <link> https://api.test.com/getnzb/1efe314025c6661380c7edf9938c38b3.nzb&amp;i=341878&amp;r=TEST</link>
      <comments>https://api.test.com/details/1efe314025c6661380c7edf9938c38b3#comments</comments>
      <pubDate>Sun, 28 Apr 2019 11:01:32 -0400</pubDate>
      <category>Audio &gt; MP3</category>
      <description>Test Test</description>
      <enclosure url="https://api.test.com/getnzb/1efe314025c6661380c7edf9938c38b3.nzb&amp;i=341878&amp;r=TEST" length="174348576" type="application/x-nzb" />
      <newznab:attr name="category" value="3000" />
      <newznab:attr name="category" value="3010" />
      <newznab:attr name="size" value="174348576" />
      <newznab:attr name="guid" value="1efe314025c6661380c7edf9938c38b3" />
      <newznab:attr name="hash" value="19e1499b8460797e1c1e391b02dfde10" />
    </item>
  </channel>
</rss>
`

func TestRSSFeedMarshal_RoundTrips(t *testing.T) {

	var v newznab.RssFeed
	err := xmlutil.Unmarshal([]byte(testXml), &v)
	assert.Nil(t, err)

	res, err := xmlutil.Marshal(v)
	assert.Nil(t, err)

	var rt newznab.RssFeed
	err = xmlutil.Unmarshal(res, &rt)
	assert.Nil(t, err)

	assert.EqualValues(t, v, rt)
}

// Because golang always marshals empty tags as <element></element>
// we need to regex replace this
var emptyTagRegexp = regexp.MustCompile("></.+?>")

func TestRSSFeedMarshal_MarshalledExactlyEqual(t *testing.T) {

	var v newznab.RssFeed
	err := xmlutil.Unmarshal([]byte(testXml), &v)
	assert.Nil(t, err)

	res, err := xmlutil.MarshalIndent(v, "", "  ")
	assert.Nil(t, err)

	stripped := emptyTagRegexp.ReplaceAllString(string(res), " />")
	assert.EqualValues(t, strings.TrimSpace(testXml), stripped)
}
