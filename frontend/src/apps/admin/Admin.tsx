import { Admin, Resource } from "react-admin";
import { CustomRoutes } from "ra-core";
import SlugRedirect from "./components/SlugRedirect";
import { BrowserRouter, Route } from "react-router";
import { MyTheme, MyDarkTheme } from "./theme/MyTheme";
import { MyLayout } from "./theme/MyLayout";
import { dataProvider } from "./providers/dataProvider";
import { authProvider } from "./providers/authProvider";
import { OidcLogin } from "./components/auth/OidcLogin";
import { useAppConfig } from "@core/config/useConfig";
import organizers from "./resources/organizer";
import eventYears from "./resources/event-year";
import teams from "./resources/team";
import invites from "./resources/invite";
import codeOfConducts from "./resources/code-of-conduct";

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
      authProvider={authProvider}
      theme={MyTheme}
      darkTheme={MyDarkTheme}
      layout={MyLayout}
      title="Crewapparat Admin"
    >
      <Resource {...organizers} />
      <Resource {...eventYears} />
      <Resource {...teams} />
      <Resource {...invites} />
      <Resource {...codeOfConducts} />
      <CustomRoutes>
        <Route path="/u/:slug" element={<SlugRedirect />} />
      </CustomRoutes>
    </Admin>
  );
};
