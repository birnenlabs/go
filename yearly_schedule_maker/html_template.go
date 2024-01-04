package main

const rowTmpl = `
  <tr>
%s
  </tr>
`

const monthCellTmpl = `
    <td class='month' colspan='3'>%s</td>
`

const dayCellTmpl = `
    <td class='dow %v'>%v</td>
    <td class='dom %v'>%v</td>
    <td class='person %v'>%v</td>
`

const personTmpl = `
      <p>%d. %v</p>
`

const satClass = "saturday"
const sunClass = "sunday"

const htmlTemplateOuter = `
<!DOCTYPE html>
<html>
<head>
<title>Waschplan</title>
<meta charset="utf-8" />
<style type='text/css'>
  body {
	  font-family: "Liberation Sans", sans;
  }
  .header {
    width: 28cm;
  }
  .year {
    display: inline-block;
    font-size: 40pt;
    font-weight: bold;
    height: 45pt;
    width: 15%%;
  }
  .person-list {
    display: inline-flex;
    flex-direction: column;
    flex-wrap: wrap;
    height: 40pt;
    vertical-align: middle;
    width: 70%%;
  }
  .person-list p {
    display: inline-block;
    margin: 0px;
  }
  .desc {
    display: inline-block;
    float: right;
    font-size: 8pt;
    line-height: 130%%;
    text-align: right;
    width: 10%%;
  }

  .schedule {
    border-collapse: collapse;
    border: 1px solid black;
    font-size: 12pt;
    table-layout: fixed;
    width: 28cm;
  }
  .schedule tr:first-child {
    border-bottom: 1px solid black;
  }
  .month {
    font-weight: bold;
    padding: 5px 2px;
    text-align: center;
    vertical-align: middle;
  }
  .dow {
    padding-left: 5px;
  }
  .dow, .dom {
    color: #777;
  }
  .person {
    font-weight: bold;
    text-align: center;
  }
  .dow, .month {
    border-left: 1px solid black;
  }
  .person, .month {
    border-right: 1px solid black;
  }
  .saturday {
    background-color: #cdf;
    color: #777;
    font-weight: normal;
  }
  .sunday {
    background-color: #fcc;
    color: red;
  }

</style>
</head>
<body>
  <div class='header'>
    <div class='year'>%d</div>
    <div class='person-list'>
  %s
    </div>
    <div class='desc'>
      Online:<br>
      %s
    </div>
  </div>

<table class='schedule'>
%s
</table>
</body>
</html>
`
