import { Service } from "./bindings/changeme";
import { Events } from "@wailsio/runtime";

const resultElement = document.getElementById("result");
const timeElement = document.getElementById("time");

window.submit = () => {
  let apiKey = document.getElementById("name").value;
  Service.GetUser(apiKey)
    .then((result) => {
      if (result) {
        resultElement.innerText = result;
      } else {
        Service.HideWindow();
        Service.Run();
      }
    })
    .catch((err) => {
      console.log(err);
    });
};

window.openBacklogAPI = () => {
  Service.OpenURL("https://eng-sol.backlog.com/EditApiSettings.action");
};

Events.On("time", (time) => {
  timeElement.innerText = time.data;
});
