$(document).ready(function() {
  var main = {
    init: function() {
      this.activeTasks = [];
      this.taskLists = [];
      this.startGetResults;
      this.host;
      this.initEvents();
      this.getHost();
    },
    initEvents: function() {
      $("#select_brand").on("change", $.proxy(this.getProfiles, this));
      $("#upload_profile").on("change", $.proxy(this.uploadProfile, this));
      $("#upload_csv").on("change", $.proxy(this.uploadCsv, this));
      $("#upload_payload").on("change", $.proxy(this.uploadPayload, this));
      $("#profiles_list").on("change", $.proxy(this.getTaskLists, this));
      $("#run_submit").on("click", $.proxy(this.submitTasks, this));
      $("#stop_submit").on("click", $.proxy(this.stopTasks, this));
      $("#reset_all").on("click", $.proxy(this.resetAll, this));
      $("#update_submit").on("click", $.proxy(this.updateRate, this));
      $("#task_lists")
        .multiselect({
          allSelectedText: "All",
          maxHeight: 200,
          includeSelectAllOption: true
        })
        .multiselect("updateButtonText");
      $("#active_task_lists")
        .multiselect({
          allSelectedText: "All",
          maxHeight: 200,
          includeSelectAllOption: true
        })
        .multiselect("updateButtonText");
    },
    updateDropDowns: function(nodeId, val, type) {
      var data = [];
      if (type === "multi") {
        for (var i = 0; i < val.length; i++) {
          obj = JSON.parse('{"label": "'+ val[i] +'", "title": "' + val[i] + '", "value": "' + val[i] +'"}')
          data.push(obj)
        }
        $("#" + nodeId).multiselect("dataprovider", data);
        $("#" + nodeId).multiselect("rebuild");
      } else {
          $("#" + nodeId).empty();
        $("#" + nodeId).append(
          "<option value=" + "" + "> Select </option>");
        $.map(val, function(x) {
          return $("#" + nodeId).append(
            "<option value=" + x + ">" + x + "</option>"
          );
        });
      }
    },
    getSession: function() {
      var self = this;
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/getsession",
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          if(response.length > 0){
            $("#select_brand")
              .parent()
              .removeClass("d-none");
              this.updateDropDowns("select_brand",response[0], "single");
          //  $("#select_brand").append(
          //      "<option value=" + response[0] + ">" + response[0] + "</option>");
            $("#select_brand option[value="+response[0][0]+"]").attr('selected', 'selected');
            $("#select_brand").attr('disabled','disabled');

            $("#profiles_list")
              .parent()
              .removeClass("d-none");
            //$("#profiles_list").append(
            //    "<option value=" + response[1] + ">" + response[1] + "</option>");
            this.updateDropDowns("profiles_list", response, "single");
            $("#profiles_list option[value="+response[1][0]+"]").attr('selected', 'selected');
            $("#profiles_list").attr('disabled','disabled');


            $("#task_lists, #run_submit")
              .parents("li")
              .removeClass("d-none");

            this.updateDropDowns("task_lists", response[2], "multi");

            $("#active_task_lists, #stop_submit")
              .parents("li")
              .removeClass("d-none");

            this.updateDropDowns("active_task_lists", response[2], "multi");


            self.getActiveTasks();

          } else {
            self.getBranches();
          }
        }, this)
      });
    },
    getBranches: function() {
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/getbranches",
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          this.updateDropDowns("select_brand",response, "single");
          $("#select_brand")
            .parent()
            .removeClass("d-none");
        }, this)
      });
    },
    getHost: function() {
      var self = this;
      $.ajax({
        url: "http://hudor.perfus.swiggyops.de:8080/api/v1/hudor/gethost",
        //url: "http://localhost:8080/api/v1/hudor/gethost",
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          console.log(response[0]);
          self.host=response[0];
          //self.host="localhost";
          self.getSession();
        }, this)
      });
    },
    getProfiles: function(e) {
      var $this = $(e.currentTarget);
      var selectedBranch = $("#select_brand").val();
      $this.attr("disabled", "disabled");
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/getprofiles",
        dataType: "json",
        method: "GET",
        data: { selectedBranch: selectedBranch },
        success: $.proxy(function(response) {
          this.updateDropDowns("profiles_list", response, "single");
          $("#upload_profile, #upload_profile_form, #upload_or")
              .parent()
              .removeClass("d-none");
          $("#profiles_list")
            .parent()
            .removeClass("d-none");
        }, this)
      });
    },
    refreshProfiles: function(e) {
      var $this = $(e.currentTarget);
      var selectedBranch = $("#select_brand").val();
      //$this.attr("disabled", "disabled");
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/getprofiles",
        dataType: "json",
        method: "GET",
        data: { selectedBranch: selectedBranch },
        success: $.proxy(function(response) {
          this.updateDropDowns("profiles_list", response, "single");
        }, this)
      });
    },
    uploadProfile: function(e) {
      event.preventDefault();
      var self = this;
      // Get form
      var form = $('#upload_profile_form')[0];
      // Create an FormData object
      var data = new FormData(form);
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/upload/task",
        type: "POST",
        processData: false,
        enctype: 'multipart/form-data',
        processData: false,
        contentType: false,
        cache: false,
        timeout: 600000,
        data: data,
        success: $.proxy(function(response) {
          if (response === "200") {
            $("#upload_notify .alert-info").removeClass("d-none");
          } else {
            $("#upload_notify .alert-danger").removeClass("d-none");
          }
          setTimeout(function() {
            $("#upload_notify .alert").addClass("d-none");
          }, 5000)
          self.refreshProfiles(e);
        }, this)
      });
    },
    uploadCsv: function(e) {
      event.preventDefault();
      var self = this;
      // Get form
      var form = $('#upload_csv_form')[0];
      // Create an FormData object
      var data = new FormData(form);
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/upload/csv",
        type: "POST",
        processData: false,
        enctype: 'multipart/form-data',
        processData: false,
        contentType: false,
        cache: false,
        timeout: 600000,
        data: data,
        success: $.proxy(function(response) {
          if (response === "200") {
            $("#upload_notify_csv .alert-info").removeClass("d-none");
          } else {
            $("#upload_notify_csv .alert-danger").removeClass("d-none");
          }
          setTimeout(function() {
            $("#upload_notify_csv .alert").addClass("d-none");
          }, 5000)
        }, this)
      });
    },
    uploadPayload: function(e) {
      event.preventDefault();
      var self = this;
      // Get form
      var form = $('#upload_payload_form')[0];
      // Create an FormData object
      var data = new FormData(form);
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/upload/payload",
        type: "POST",
        processData: false,
        enctype: 'multipart/form-data',
        processData: false,
        contentType: false,
        cache: false,
        timeout: 600000,
        data: data,
        success: $.proxy(function(response) {
          if (response === "200") {
            $("#upload_notify_payload .alert-info").removeClass("d-none");
          } else {
            $("#upload_notify_payload .alert-danger").removeClass("d-none");
          }
          setTimeout(function() {
            $("#upload_notify_payload .alert").addClass("d-none");
          }, 5000)
        }, this)
      });
    },
    getTaskLists: function(e) {
      var $this = $(e.currentTarget);
      var selectedProfile = $this.val();
      $this.attr("disabled", "disabled");
      var self = this;
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/getalltasks/" + selectedProfile,
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          self.getLogs();
          $("#upload_profile, #upload_or")
              .parents("li")
              .addClass("d-none");
          $("#task_lists, #run_submit")
            .parents("li")
            .removeClass("d-none");

          this.updateDropDowns("task_lists", response, "multi");
          $("#upload_payload, #upload_payload_form")
              .parent()
              .removeClass("d-none");
          $("#upload_csv, #upload_csv_form")
              .parent()
              .removeClass("d-none");

          $("#active_task_lists, #stop_submit")
            .parents("li")
            .removeClass("d-none");

          this.updateDropDowns("active_task_lists", response, "multi");
        }, this)
      });
    },
    submitTasks: function(e) {
      var selectedTasks = $("#task_lists").val();
      var self = this;
      var tasks = "";
      for (var i = 0; i < selectedTasks.length; i++) {
        tasks = tasks + selectedTasks[i] + ","
      }
      tasks = tasks.slice(0, -1);
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/runtask/"+ tasks,
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          self.getActiveTasks(response);
        }, this)
      });
    },
    stopTasks: function(e) {
      var selectedTasks = $("#active_task_lists").val();
      var self = this;
      var tasks = "";
      for (var i = 0; i < selectedTasks.length; i++) {
        tasks = tasks + selectedTasks[i] + ","
      }
      tasks = tasks.slice(0, -1);
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/stopTask/"+ tasks,
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          if (response === "200") {
            $("#stop_notify .alert-info").removeClass("d-none");
          } else {
            $("#stop_notify .alert-danger").removeClass("d-none");
          }
          setTimeout(function() {
            $("#stop_notify .alert").addClass("d-none");
          }, 5000)
        }, this)
      });
    },
    updateRate: function() {
      var activeTask = $("#active_task").val();
      var rateValue = $("#rate_value").val();
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/updateTaskRate/"+activeTask+"/"+rateValue,
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          if (response === "200") {
            $("#notify .alert-info").removeClass("d-none");
          } else {
            $("#notify .alert-danger").removeClass("d-none");
          }
          setTimeout(function() {
            $("#notify .alert").addClass("d-none");
          }, 5000);
        }, this)
      });
    },
    getActiveTasks: function() {
      var self = this
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/getActiveTasks",
        dataType: "json",
        method: "GET",
        // data: { selectedBrand: selectedTasks },
        success: $.proxy(function(response) {
          console.log(response);
          $("#active_task")
              .parents("li")
              .removeClass("d-none");
          this.updateDropDowns("active_task", response, "single");
          $("#update_block").removeClass("d-none");
          if(response.length > 0){
            self.getTasksResults();
            this.startGetResults = setInterval(function() {
              self.getTasksResults(response);
              self.getLogs();
            }, 10000);
          } else {
            console.log("clearning results");
            clearInterval(this.startGetResults);
          }

        }, this)
      });
    },
    getLogs: function() {
      var self = this
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/logs",
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          if (response.length > 0) {
            var previousValue = $("#logs").val();
            $("#logs").html(previousValue + response.toString());
          }
        }, this)
      });
    },
    getTasksResults: function(response) {
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/getresults",
        dataType: "json",
        method: "GET",
        // data: { selectedBrand: selectedTasks },
        success: $.proxy(function(response) {
          if(response.length > 0){
            this.paintResults(response);
          } else {
              $("#results_card").empty()
              clearInterval(this.startGetResults);
          }

        }, this)
      });
    },
    paintResults: function(response) {
      var self = this;
      var $parent = $("#results_card");
      $parent.empty();
      setTimeout(
        $.proxy(function() {
          response.forEach(function(val) {
            val = JSON.parse(val)
            $parent.append(self.getMarkUp(val));
          });
        }, this),
        500
      );

      $("#empty_search").addClass("d-none");
      $parent.removeClass("d-none");
    },
    getSelectedArray: function() {},
    getMarkUp: function(res) {
      return (
        '<div class="col-md-6 col-xl-3 mb-4"><div class="border-left-primary card h-100 py-2 shadow"><div class=card-body><div class="align-items-center no-gutters row"><div class="col mr-2 text-center"><div class="font-weight-bold h5 mb-1 text-primary text-sm text-uppercase">' +
        res.title +
        '</div><br></div></div><div class="align-items-center no-gutters row"><div class="col mr-2 text-center"><div class="font-weight-bold mb-1 text-sm text-uppercase text-success">Success</div><div class="font-weight-bold h5 mb-0 text-gray-800">' +
        res.success +
        '</div></div><div class="col mr-2 text-center"><div class="font-weight-bold mb-1 text-sm text-uppercase text-warning">Rate</div><div class="font-weight-bold h5 mb-0 text-gray-800">' +
        res.rate +
        '</div></div><div class="col mr-2 text-center"><div class="font-weight-bold mb-1 text-sm text-uppercase text-danger">ERROR</div><div class="font-weight-bold h5 mb-0 text-gray-800">' +
        res.error +
        "</div></div></div></div></div></div>"
      );
    },
    resetAll: function() {
      var self = this;
      $.ajax({
        url: "http://"+this.host+":8080/api/v1/hudor/reset",
        dataType: "json",
        method: "GET",
        success: $.proxy(function(response) {
          clearInterval(this.startGetResults);
          location.reload("forcedReload");
        }, this)
      });
    }
  };
  main.init();
});
