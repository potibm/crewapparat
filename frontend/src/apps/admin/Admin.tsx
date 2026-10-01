import { Admin } from "react-admin";
import { BrowserRouter } from "react-router";
import { MyTheme, MyDarkTheme } from "./theme/MyTheme";
import { MyLayout } from "./theme/MyLayout";
import { dataProvider } from "./providers/dataProvider";
import { authProvider } from "./providers/authProvider";
import { OidcLogin } from "./components/auth/OidcLogin";
import { useAppConfig } from "@core/config/useConfig";

export const AdminApp = () => (
  <BrowserRouter>
    <AppBootstrapper />
  </BrowserRouter>
);

export const AppBootstrapper = () => {
  const appConfig = useAppConfig();
  const isOidcActive = appConfig.auth?.type === "oidc";

  return (
    <Admin
      loginPage={isOidcActive ? OidcLogin : false}
      dataProvider={dataProvider}
      authProvider={isOidcActive ? authProvider : undefined}
      theme={MyTheme}
      darkTheme={MyDarkTheme}
      layout={MyLayout}
      title="Crewapparat Admin"
    />
  );
};
