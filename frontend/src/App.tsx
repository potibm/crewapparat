import * as Sentry from "@sentry/react";
import { AdminApp } from "@admin/Admin";
import InvitePage from "@public/InvitePage";

function App() {
  if (window.location.pathname.startsWith("/invite/")) {
    return (
      <div className="App">
        <Sentry.ErrorBoundary
          fallback={
            <p>A serious error has occurred. Please reload the application.</p>
          }
        >
          <InvitePage />
        </Sentry.ErrorBoundary>
      </div>
    );
  }

  return (
    <div className="App">
      <Sentry.ErrorBoundary
        fallback={
          <p>A serious error has occurred. Please reload the application.</p>
        }
      >
        <AdminApp />
      </Sentry.ErrorBoundary>
    </div>
  );
}

export default App;
