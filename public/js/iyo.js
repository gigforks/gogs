function searchOrgs() {
    if (!$('#search-org-box .results').length) {
        return;
    }

    var $searchOrgBox = $('#search-org-box');
    var $results = $searchOrgBox.find('.results');
    $searchOrgBox.keyup(function () {
        var $this = $(this);
        var keyword = $this.find('input').val();
        if (keyword.length < 2) {
            $results.hide();
            return;
        }

        $.ajax({
            url: suburl + '/api/v1/orgs/search?q=' + keyword,
            dataType: "json",
            success: function (response) {
                var notEmpty = function (str) {
                    return str && str.length > 0;
                };

                $results.html('');

                if (response.ok && response.data.length) {
                    var html = '';
                    $.each(response.data, function (i, item) {
                        html += '<div class="item"><span class="username">' + item.Name + '</span>' + '</div>';
                    });
                    $results.html(html);
                    $this.find('.results .item').click(function () {
                        $this.find('input').val($(this).find('.username').text());
                        $results.hide();
                    });
                    $results.show();
                } else {
                    $results.hide();
                }
            }
        });
    });
    $searchOrgBox.find('input').focus(function () {
        $searchOrgBox.keyup();
    });
    hideWhenLostFocus('#search-org-box .results', '#search-org-box');
}

$(document).ready(function () {
    searchOrgs();
});
